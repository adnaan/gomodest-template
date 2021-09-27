const opList = "list";
const opDelete = "delete";
const opInsert = "insert";
const opUpdate = "update";
const opGet = "get";
let methodID = 0;

const call = (method, params) => {
    methodID += 1
    return {
        jsonrpc: "2.0",
        method: method,
        id: methodID,
        params: params
    }
}

const method = (resource, op) => {
    return resource ? `${resource}/${op}` : op
}

export {
    opList,
    opInsert,
    opUpdate,
    opGet,
    opDelete,
    method,
    call
}