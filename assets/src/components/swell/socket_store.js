import {writable} from "svelte/store";
import isEqual from "lodash.isequal";

const reopenTimeouts = [2000, 5000, 10000, 30000, 60000];

const jsonRPC2Message = (method, params, id) => {
    if (id) {
        id = `${method}:${id}`
    }
    return {
        jsonrpc: "2.0",
        method: method,
        id: id,
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
            const {subscribe: subscribeLoaders, set: setLoaders, update: updateLoaders} = writable({});
            const {subscribe: subscribeErrors, set: setErrors, update: updateErrors} = writable({});
            const changeEvents = Object.keys(changeEventHandlers);
            const messageHandler = (message) => {
                if (isEqual(message, initialValue)) {
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

                let method = message.id;
                let id = message.id;
                if (method.includes(":")) {
                    const parts = method.split(":");
                    if (parts.length === 2) {
                        method = parts[0];
                        id = parts[1];
                    }
                }

                if (message.error) {
                    if (changeEvents.includes('error')) {
                        changeEventHandlers['error'](undefined, message.error)
                    }
                    updateErrors(errors => {
                        let newError = {}
                        newError[id] = message.error;
                        return {...errors, ...newError}
                    })
                    return;
                }

                if (!message.result) {
                    if (changeEvents.includes('error')) {
                        changeEventHandlers['error'](undefined, 'result is undefined')
                    }

                    updateErrors(errors => {
                        let newError = {}
                        newError[id] = {message: 'result is undefined'};
                        return {...errors, ...newError}
                    })
                    return;
                }

                updateLoaders(loaders => {
                    delete loaders[id];
                    loaders = loaders
                    return loaders
                })

                const handler = changeEventHandlers[method]
                update((data) => handler(data, message.result))
            }
            if (changeEventPrefix) {
                prefixedSubscriptions.set(changeEventPrefix, messageHandler);
            } else {
                subscriptions.add(messageHandler)
            }

            return {
                subscribe,
                change: (changeEvent, params, id) => {
                    if (!changeEvent){
                        throw 'changeEvent is required';
                    }
                    const message = jsonRPC2Message(changeEvent, params, id);
                    const send = () => socket.send(JSON.stringify(message));
                    if (!socket || socket && socket.readyState !== WebSocket.OPEN) openSocket().then(send);
                    else send();
                    // set loading true for current pending operation
                    let loader = {}
                    loader[id ? id: changeEvent] = true
                    console.log(loader);
                    updateLoaders(loaders => {return {...loaders,...loader}})
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
                },
                loaders: () => {
                    return {
                        subscribe:subscribeLoaders,
                        reset: (id) => {
                            if(id){
                                updateLoaders(loaders => {
                                    delete loaders[id];
                                    loaders = loaders;
                                    return loaders
                                })
                            } else {
                                setLoaders({});
                            }

                        }
                    }
                },
                errors: () => {
                    return {
                        subscribe: subscribeErrors,
                        reset: (id) => {
                            if(id){
                                updateErrors(errors => {
                                    delete errors[id];
                                    errors = errors;
                                    return errors
                                })
                            } else {
                                setErrors({});
                            }

                        }
                    }
                }
            }
        }
    }
}

export {
    createJsonrpc2Socket,
}