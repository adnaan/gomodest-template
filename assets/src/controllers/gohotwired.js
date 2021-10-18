import * as Turbo from "@hotwired/turbo"

const reopenTimeouts = [2000, 5000, 10000, 30000, 60000];

const createEventDispatcher = (url, socketOptions) => {
    let socket, openPromise, reopenTimeoutHandler;
    let reopenCount = 0;

    // socket code copied from https://github.com/arlac77/svelte-websocket-store/blob/master/src/index.mjs
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
        reopenTimeoutHandler = setTimeout(() =>
                openSocket().then(() => Turbo.session.connectStreamSource(socket)),
            reopenTimeout());
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

    openSocket().then(() => {
        Turbo.session.connectStreamSource(socket);
    });
    return (eventID, action, target, content, params) => {
        if (!eventID) {
            throw 'eventID is required';
        }
        const event = {
            id: eventID,
            action: action,
            target: target,
            content_template: content,
            params: params
        }
        const send = () => socket.send(JSON.stringify(event));
        if (!socket || socket && socket.readyState !== WebSocket.OPEN) openSocket().then(send);
        else send();
    }
}

export {
    createEventDispatcher,
}