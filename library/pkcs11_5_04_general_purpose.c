/* This file is auto-generated from pkcs11_5_04_general_purpose.rpc by rpcc. */
/* -*- c -*-
 *
 * Copyright (C) 2020-2021 Markku Rossi.
 *
 * All rights reserved.
 */

#include "vp_includes.h"

/** Version: 3.0 */
/** Section: 5.4 General-purpose functions */

#define CK_PKCS11_FUNCTION_INFO(name) name,

static struct CK_FUNCTION_LIST_3_0 function_list =
  {
    {
      CRYPTOKI_VERSION_MAJOR,
      CRYPTOKI_VERSION_MINOR,
    },

#include "pkcs11f.h"

  };

static CK_C_INITIALIZE_ARGS init_args;

static CK_RV
mutex_create(void **ret)
{
  pthread_mutex_t *mutex;

  if (ret == NULL)
    return CKR_ARGUMENTS_BAD;

  mutex = calloc(1, sizeof(*mutex));
  if (mutex == NULL)
    return CKR_HOST_MEMORY;

  if (pthread_mutex_init(mutex, NULL) != 0)
    {
      free(mutex);
      return CKR_HOST_MEMORY;
    }

  *ret = mutex;

  return CKR_OK;
}

static CK_RV
mutex_destroy(void *ptr)
{
  if (ptr != NULL)
    {
      pthread_mutex_t *mutex = (pthread_mutex_t *) ptr;

      pthread_mutex_destroy(mutex);
      free(mutex);
    }

  return CKR_OK;
}

static CK_RV
mutex_lock(void *ptr)
{
  pthread_mutex_t *mutex = (pthread_mutex_t *) ptr;

  if (mutex == NULL)
    return CKR_ARGUMENTS_BAD;

  if (pthread_mutex_lock(mutex) != 0)
    return CKR_CANT_LOCK;

  return CKR_OK;
}

static CK_RV
mutex_unlock(void *ptr)
{
  pthread_mutex_t *mutex = (pthread_mutex_t *) ptr;

  if (mutex == NULL)
    return CKR_ARGUMENTS_BAD;

  if (pthread_mutex_unlock(mutex) != 0)
    return CKR_ARGUMENTS_BAD;

  return CKR_OK;
}

/* C_Initialize initializes the Cryptoki library. */
CK_RV
C_Initialize
(
  CK_VOID_PTR   pInitArgs  /* if this is not NULL_PTR, it gets
                            * cast to CK_C_INITIALIZE_ARGS_PTR
                            * and dereferenced
                            */
)
{
  VPBuffer buf;
  unsigned char *data;
  size_t len;

  if (pInitArgs != NULL)
    {
      memcpy(&init_args, pInitArgs, sizeof(init_args));
    }
  else
    {
      memset(&init_args, 0, sizeof(init_args));
      init_args.CreateMutex = mutex_create;
      init_args.DestroyMutex = mutex_destroy;
      init_args.LockMutex = mutex_lock;
      init_args.UnlockMutex = mutex_unlock;
    }


  /* XXX use global session */

  vp_buffer_init(&buf);
  vp_buffer_add_uint32(&buf, 0xc0050401);
  vp_buffer_add_space(&buf, 4);

  data = vp_buffer_ptr(&buf);
  if (data == NULL)
    {
      vp_buffer_uninit(&buf);
      return CKR_HOST_MEMORY;
    }
  len = vp_buffer_len(&buf);
  VP_PUT_UINT32(data + 4, len - 8);


  return CKR_OK;
}

/* C_Finalize indicates that an application is done with the
 * Cryptoki library.
 */
CK_RV
C_Finalize
(
  CK_VOID_PTR   pReserved  /* reserved.  Should be NULL_PTR */
)
{
  VP_FUNCTION_ENTER;
  return CKR_OK;
}

/* C_GetInfo returns general information about Cryptoki. */
CK_RV
C_GetInfo
(
  CK_INFO_PTR   pInfo  /* location that receives information */
)
{
  VP_FUNCTION_NOT_SUPPORTED;
}

/* C_GetFunctionList returns the function list. */
CK_RV
C_GetFunctionList
(
  CK_FUNCTION_LIST_PTR_PTR ppFunctionList  /* receives pointer to
                                            * function list
                                            */
)
{
  VP_FUNCTION_ENTER;

  if (ppFunctionList == NULL)
    return CKR_ARGUMENTS_BAD;

  *ppFunctionList = (CK_FUNCTION_LIST_PTR) &function_list;

  return CKR_OK;
}

/* C_GetInterfaceList returns all the interfaces supported by the module*/
CK_RV
C_GetInterfaceList
(
  CK_INTERFACE_PTR  pInterfacesList,  /* returned interfaces */
  CK_ULONG_PTR      pulCount          /* number of interfaces returned */
)
{
  VP_FUNCTION_NOT_SUPPORTED;
}

/* C_GetInterface returns a specific interface from the module. */
CK_RV
C_GetInterface
(
  CK_UTF8CHAR_PTR       pInterfaceName, /* name of the interface */
  CK_VERSION_PTR        pVersion,       /* version of the interface */
  CK_INTERFACE_PTR_PTR  ppInterface,    /* returned interface */
  CK_FLAGS 		flags           /* flags controlling the semantics
                                         * of the interface */
)
{
  VP_FUNCTION_NOT_SUPPORTED;
}
