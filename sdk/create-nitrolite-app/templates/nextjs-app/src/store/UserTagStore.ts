import { proxy } from 'valtio';

export interface IUserTagState {
    userTag: string | null;
    loading: boolean;
    error: string | null;
}

const state = proxy<IUserTagState>({
    userTag: null,
    loading: false,
    error: null,
});

const UserTagStore = {
    state,
    setUserTag(userTag: string): void {
        state.userTag = userTag;
        state.loading = false;
        state.error = null;
    },
    setLoading(loading: boolean): void {
        state.loading = loading;
    },
    setError(error: string | null): void {
        state.error = error;
        state.loading = false;
    },
    clearUserTag(): void {
        state.userTag = null;
        state.error = null;
    },
};

export default UserTagStore;
