// @ts-nocheck
/**
 * Generated by orval v7.8.0 🍺
 * Do not edit manually.
 * Inbox Whisperer API
 * OpenAPI specification for the Inbox Whisperer backend API. This spec will be expanded as endpoints are implemented.

 * OpenAPI spec version: 0.1.0
 */
import axios from 'axios';
import type {
  AxiosRequestConfig,
  AxiosResponse
} from 'axios';

import type {
  User,
  UserCreateRequest,
  UserUpdateRequest
} from '../inboxWhispererAPI.schemas';




  export const getUser = () => {
/**
 * Only admin can list users. Non-admins receive 403 Forbidden.
 * @summary List users
 */
const getUsers = <TData = AxiosResponse<User[]>>(
     options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.get(
      `/users`,options
    );
  }
/**
 * Only admin can create users. Non-admins receive 403 Forbidden.
 * @summary Create a user
 */
const postUsers = <TData = AxiosResponse<User>>(
    userCreateRequest: UserCreateRequest, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.post(
      `/users`,
      userCreateRequest,options
    );
  }
/**
 * Only the user themselves (or admin) can get this user. Others receive 403 Forbidden.
 * @summary Get user by ID
 */
const getUsersId = <TData = AxiosResponse<User>>(
    id: string, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.get(
      `/users/${id}`,options
    );
  }
/**
 * Only the user themselves (or admin) can update this user. Others receive 403 Forbidden.
 * @summary Update user
 */
const putUsersId = <TData = AxiosResponse<User>>(
    id: string,
    userUpdateRequest: UserUpdateRequest, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.put(
      `/users/${id}`,
      userUpdateRequest,options
    );
  }
/**
 * Only the user themselves (or admin) can delete this user. Others receive 403 Forbidden.
 * @summary Delete user
 */
const deleteUsersId = <TData = AxiosResponse<void>>(
    id: string, options?: AxiosRequestConfig
 ): Promise<TData> => {
    return axios.delete(
      `/users/${id}`,options
    );
  }
return {getUsers,postUsers,getUsersId,putUsersId,deleteUsersId}};
export type GetUsersResult = AxiosResponse<User[]>
export type PostUsersResult = AxiosResponse<User>
export type GetUsersIdResult = AxiosResponse<User>
export type PutUsersIdResult = AxiosResponse<User>
export type DeleteUsersIdResult = AxiosResponse<void>
