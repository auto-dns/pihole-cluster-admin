import { FullInitStatus } from "../types";

export function isFullyInitialized(status?: FullInitStatus): boolean {
    if (!status) return false;
    return status.userCreated && status.piholeStatus !== "UNINITIALIZED";
}
