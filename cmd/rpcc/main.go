//
// Copyright (c) 2021 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/markkurossi/pkcs11-provider/ipc"
)

var (
	reVersion      = regexp.MustCompilePOSIX(`/\*\*[[:space:]]+Version:[[:space:]]+([[:digit:]]+)\.([[:digit:]]+)`)
	reSection      = regexp.MustCompilePOSIX(`/\*\*[[:space:]]+Section:[[:space:]]+([[:digit:]]+)\.([[:digit:]]+)[[:space:]]*([^\*]+)`)
	reFunction     = regexp.MustCompilePOSIX(`^(C_[a-zA-Z0-9_]+)`)
	reSigStart     = regexp.MustCompilePOSIX(`^[[:space:]]*/\*\*[[:space:]]*$`)
	reSigEnd       = regexp.MustCompilePOSIX(`^[[:space:]]*\*/[[:space:]]*$`)
	reFieldSection = regexp.MustCompilePOSIX(`^[[:space:]]*\*[[:space:]]*([[:alnum:]]+):[[:space:]]*$`)
	reField        = regexp.MustCompilePOSIX(`\*[[:space:]]+([[:^space:]]+)[[:space:]]+([[:^space:]]+)`)
	reType         = regexp.MustCompilePOSIX(`^(\[([A-Za-z_0-9]+)\])?([A-Za-z_0-9]+)$`)

	vMajorLast int
	vMinorLast int

	output     io.WriteCloser = os.Stdout
	outputC    bool
	outputGo   bool
	printInfo  bool
	printDebug bool
)

func main() {
	log.SetFlags(0)
	flag.BoolVar(&outputC, "c", false, "generate C code")
	flag.BoolVar(&printInfo, "i", false, "print file information")
	flag.BoolVar(&printDebug, "d", false, "print debug information")
	typeFile := flag.String("t", "", "type information file")
	o := flag.String("o", "", "output file name")
	flag.Parse()

	fmt.Printf("PKCS #11 RPC Compiler\n")

	var err error

	if len(*typeFile) > 0 {
		err = readTypes(*typeFile)
		if err != nil {
			log.Fatalf("failed to read types: %s\n", err)
		}
	}

	if len(*o) != 0 {
		output, err = os.Create(*o)
		if err != nil {
			log.Fatalf("failed to create output file: %s", err)
		}
		defer output.Close()
	}

	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			log.Fatalf("%s: %s\n", arg, err)
		}
		defer f.Close()
		header(arg)
		err = processFile(f)
		if err != nil {
			log.Fatalf("%s: %s\n", arg, err)
		}
	}
}

func processFile(in io.Reader) error {
	reader := bufio.NewReader(in)

	var vMajor, vMinor, s0, s1, s2 int
	var title string

	// Parse header.
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		print(line)
		m := reVersion.FindStringSubmatch(line)
		if m != nil {
			vMajori64, err := strconv.ParseInt(m[1], 10, 8)
			if err != nil {
				return err
			}
			vMajor = int(vMajori64) - 2

			vMinori64, err := strconv.ParseInt(m[2], 10, 8)
			if err != nil {
				return err
			}
			vMinor = int(vMinori64)
			continue
		}
		m = reSection.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		s0i64, err := strconv.ParseInt(m[1], 10, 8)
		if err != nil {
			return err
		}
		s0 = int(s0i64)

		s1i64, err := strconv.ParseInt(m[2], 10, 8)
		if err != nil {
			return err
		}
		s1 = int(s1i64)

		title = strings.TrimSpace(m[3])
		break
	}
	if vMajor != vMajorLast || vMinor != vMinorLast {
		vMajorLast = vMajor
		vMinorLast = vMinor
		info("* Version %d.%d\n", vMajor+2, vMinor)
	}
	info("** %d.%d %s\n", s0, s1, title)

	// Process all functions
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		m := reFunction.FindStringSubmatch(line)
		if m != nil {
			print(line)
			s2++
			name := m[1]
			info(" - %d.%d.%d %s\n", s0, s1, s2, name)
			continue
		}
		m = reSigStart.FindStringSubmatch(line)
		if m == nil {
			print(line)
			continue
		}
		// Process signature

		var inputs []Field
		var outputs []Field
		var fieldSection *[]Field

		for {
			line, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			if reSigEnd.FindStringSubmatch(line) != nil {
				// Signature processed
				break
			}
			m = reFieldSection.FindStringSubmatch(line)
			if m != nil {
				switch m[1] {
				case "Inputs":
					fieldSection = &inputs

				case "Outputs":
					fieldSection = &outputs

				default:
					return fmt.Errorf("unknown field section: %s", m[1])
				}
				continue
			}

			m = reField.FindStringSubmatch(line)
			if m != nil {
				field, err := parseField(m[1], m[2])
				if err != nil {
					return err
				}
				if fieldSection == nil {
					return fmt.Errorf("field declaration without section")
				}
				*fieldSection = append(*fieldSection, field)
				continue
			}
		}

		msgType := ipc.NewType(vMajor, vMinor, s0, s1, s2)

		if outputC {
			print(`  VPBuffer buf;
  unsigned char *data;
  size_t len;
`)
			var depth int
			for _, input := range inputs {
				d, err := input.Depth()
				if err != nil {
					return err
				}
				if d > depth {
					depth = d
				}
			}
			if depth > 0 {
				print("  int")
				for i := 0; i < depth; i++ {
					if i == 0 {
						print(" ")
					} else {
						print(", ")
					}
					printf("%c", 'i'+i)
				}
				print(`;

  vp_buffer_init(&buf);
`)
			}

			printf("  vp_buffer_add_uint32(&buf, 0x%08x);\n", int(msgType))
			printf("  vp_buffer_add_space(&buf, 4);\n\n")

			for _, input := range inputs {
				err = input.Input(0)
				if err != nil {
					return err
				}
			}
			print(`
  data = vp_buffer_ptr(&buf);
  if (data == NULL)
    {
      vp_buffer_uninit(&buf);
      return CKR_HOST_MEMORY;
    }
  len = vp_buffer_len(&buf);
  VP_PUT_UINT32(data + 4, len - 8);

`)

			for idx, o := range outputs {
				fmt.Printf("  // Output %d: %#v\n", idx, o)
			}
			print("  VP_FUNCTION_NOT_SUPPORTED;\n")
		}
	}
}

func header(source string) {
	msg := fmt.Sprintf("/* This file is auto-generated from %s by rpcc. */\n",
		source)
	print(msg)
}

func info(format string, a ...interface{}) {
	if !printInfo {
		return
	}
	fmt.Printf(format, a...)
}

func debug(format string, a ...interface{}) {
	if !printDebug {
		return
	}
	fmt.Printf(format, a...)
}

func print(line string) {
	if !outputC {
		return
	}
	_, err := output.Write([]byte(line))
	if err != nil {
		log.Fatalf("write failed: %s", err)
	}
}

func printf(format string, a ...interface{}) {
	if !outputC {
		return
	}
	fmt.Fprintf(output, format, a...)
}
