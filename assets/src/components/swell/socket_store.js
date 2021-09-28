import {writable} from "svelte/store";
import isEqual from "lodash.isequal";

const reopenTimeouts = [2000, 5000, 10000, 30000, 60000];

const jsonRPC2Message = (method, params) => {
    return {
        jsonrpc: "2.0",
        method: method,
        id: method,
        params: params
    }
}

const createJsonrpc2Socket = (url, socketOptions) => {
    let socket, openPromise, reopenTimeoutHandler;
    let reopenCount = 0;
    const subscriptions = new Set();
    const prefixedSubscriptions = new Map()
    // code copied from https://github.com/arlac77/svelte-websocket-store/blob/master/src/index.mjs
    // thank you https://github.com/arlac77 !!
    function reopenTimeout() {
        const n = reopenCount;
        reopenCount++;
        return reopenTimeouts[
            n >= reopenTimeouts.length - 1 ? reopenTimeouts.length - 1 : n
            ];
    }

    function closeSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler);
        }

        if (socket) {
            socket.close();
            socket = undefined;
        }
    }

    function reOpenSocket() {
        closeSocket();
        if (subscriptions.size > 0) {
            reopenTimeoutHandler = setTimeout(() => openSocket(), reopenTimeout());
        }
    }

    async function openSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler);
            reopenTimeoutHandler = undefined;
        }

        // we are still in the opening phase
        if (openPromise) {
            return openPromise;
        }

        socket = new WebSocket(url, socketOptions);

        socket.onmessage = event => {
            const eventData = JSON.parse(event.data);
            if (!eventData.error && eventData.id) {
                let found = false;
                prefixedSubscriptions.forEach((subscription, prefix) => {
                    if (eventData.id.startsWith(prefix)) {
                        subscription(eventData);
                        found = true
                    }
                })
                // send it all subscribed without prefix
                if (!found) {
                    subscriptions.forEach(subscription => subscription(eventData));
                }
            } else {
                subscriptions.forEach(subscription => subscription(eventData));
            }

        };

        socket.onclose = event => reOpenSocket();

        openPromise = new Promise((resolve, reject) => {
            socket.onerror = error => {
                reject(error);
                openPromise = undefined;
            };
            socket.onopen = event => {
                reopenCount = 0;
                resolve();
                openPromise = undefined;
            };
        });
        return openPromise;
    }

    openSocket();
    return {
        newStore: (initialValue, changeEventHandlers, changeEventPrefix) => {
            const {subscribe, set, update} = writable(initialValue);
            const changeEvents = Object.keys(changeEventHandlers);
            const messageHandler = (message) => {
                if (isEqual(message, initialValue)) {
                    return;
                }
                if (message.error) {
                    if (changeEvents.includes('error')) {
                        changeEventHandlers['error'](undefined, message.error)
                    }
                    return;
                }
                if (!message.id) {
                    if (changeEvents.includes('error')) {
                        changeEventHandlers['error'](undefined, 'response id is undefined')
                    }
                    return;
                }
                if (!changeEvents.includes(message.id)) {
                    if (changeEvents.includes('error')) {
                        changeEventHandlers['error'](undefined, `response id ${message.id} has no changeEvent handlers`)
                    }
                    return;
                }
                if (!message.result) {
                    if (changeEvents.includes('error')) {
                        changeEventHandlers['error'](undefined, 'response is undefined')
                    }
                    return;
                }
                const handler = changeEventHandlers[message.id]
                update((data) => handler(data, message.result))
            }
            if (changeEventPrefix) {
                prefixedSubscriptions.set(changeEventPrefix, messageHandler);
            } else {
                subscriptions.add(messageHandler)
            }

            return {
                subscribe,
                change: (changeEvent, params) => {
                    const message = jsonRPC2Message(changeEvent, params);
                    const send = () => socket.send(JSON.stringify(message));
                    if (socket.readyState !== WebSocket.OPEN) openSocket().then(send);
                    else send();
                },
                close: () => {
                    if (changeEventPrefix) {
                        prefixedSubscriptions.delete(changeEventPrefix);
                    } else {
                        subscriptions.delete(messageHandler)
                    }
                    if (prefixedSubscriptions.size === 0 && subscriptions.size === 0) {
                        closeSocket();
                    }
                }
            }
        }
    }
}

export {
    createJsonrpc2Socket,
}