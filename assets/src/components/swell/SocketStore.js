import websocketStore from "svelte-websocket-store";
import {writable} from "svelte/store";
import isEqual from "lodash.isequal";

const jsonRPC2Message = (method, params) => {
    return {
        jsonrpc: "2.0",
        method: method,
        id: method,
        params: params
    }
}

const createJsonrpc2SocketStore = (conn, initialValue, changeEventHandlers) => {
    const {subscribe, set, update} = writable(initialValue)
    let socketStore = websocketStore(conn.url, conn.socketOptions);
    const changeEvents = Object.keys(changeEventHandlers);
    socketStore.subscribe((message) => {
        if (isEqual(message,initialValue)){
            return;
        }
        if (message.error) {
            if (changeEvents.includes('error')) {
                changeEventHandlers['error'](undefined,message.error)
            }
            return;
        }
        if (!message.id) {
            if (changeEvents.includes('error')) {
                changeEventHandlers['error'](undefined,'response id is undefined')
            }
            return;
        }
        if (!changeEvents.includes(message.id)) {
            if (changeEvents.includes('error')) {
                changeEventHandlers['error'](undefined,`response id ${message.id} has no changeEvent handlers`)
            }
            return;
        }
        if (!message.result) {
            if (changeEvents.includes('error')) {
                changeEventHandlers['error'](undefined,'response is undefined')
            }
            return;
        }
        const handler = changeEventHandlers[message.id]
        update((data) => handler(data, message.result))
    })


    return {
        subscribe,
        change: (changeEvent, params) => {
            socketStore.set(jsonRPC2Message(changeEvent, params));
        }
    }
}

export {
    createJsonrpc2SocketStore
}