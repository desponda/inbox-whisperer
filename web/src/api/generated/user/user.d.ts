import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import type { User, UserCreateRequest, UserUpdateRequest } from '../inboxWhispererAPI.schemas';
export declare const getUser: () => {
    getUsers: <TData = AxiosResponse<User[], any>>(options?: AxiosRequestConfig) => Promise<TData>;
    postUsers: <TData = AxiosResponse<User, any>>(userCreateRequest: UserCreateRequest, options?: AxiosRequestConfig) => Promise<TData>;
    getUsersId: <TData = AxiosResponse<User, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
    putUsersId: <TData = AxiosResponse<User, any>>(id: string, userUpdateRequest: UserUpdateRequest, options?: AxiosRequestConfig) => Promise<TData>;
    deleteUsersId: <TData = AxiosResponse<void, any>>(id: string, options?: AxiosRequestConfig) => Promise<TData>;
};
export type GetUsersResult = AxiosResponse<User[]>;
export type PostUsersResult = AxiosResponse<User>;
export type GetUsersIdResult = AxiosResponse<User>;
export type PutUsersIdResult = AxiosResponse<User>;
export type DeleteUsersIdResult = AxiosResponse<void>;
