let id = 1;
export function call(method, params) {
    id += 1
    return {
        jsonrpc: "2.0",
        method: method,
        id: id,
        params: params
    }
}
