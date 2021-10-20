import {Controller} from "@hotwired/stimulus"
import * as Turbo from "@hotwired/turbo";

export default class extends Controller {
    static values = {
        changeRequestId: String,
        action: String,
        target: String,
        content: String,
        params: Object,
    }

    initialize() {
        let todosURL = "ws://localhost:3000/samples/gh/todos"
        if (process.env.ENV === "production") {
            todosURL = `wss://${process.env.HOST}/samples/gh/todos`
        }
        this.onSocketReconnect  = () => {
            if (this.dispatcher) {
                if (!this.changeRequestIdValue || !this.actionValue || !this.targetValue || !this.contentValue) {
                    console.warn("action controller.onSocketReconnect requires changeRequestId, action, target and content params")
                    return
                }
                this.dispatcher(this.changeRequestIdValue, this.actionValue, this.targetValue, this.contentValue, this.paramsValue)
            }
        }
        this.dispatcher = changeRequestsDispatcher(todosURL, [], this.onSocketReconnect)
    }


    submit(e) {
        e.preventDefault()
        const {changeRequestId, action, target, content, ...rest} = e.params
        if (!changeRequestId || !action || !target || !content) {
            console.error("action submit requires changeRequestId, action, target and content params")
            return
        }
        let json = {...rest};
        let formData = new FormData(e.currentTarget);
        formData.forEach((value, key) => json[key] = value);
        if (this.dispatcher) {
            this.dispatcher(changeRequestId, action, target, content, json)
        }
    }

    change(e) {
        e.preventDefault()
        const {changeRequestId, action, target, content, ...rest} = e.params
        if (!changeRequestId || !action || !target || !content) {
            console.error("action change requires changeRequestId, action, target and content params")
            return
        }
        let json = {...rest};
        if (this.dispatcher) {
            this.dispatcher(changeRequestId, action, target, content, json)
        }
    }

}

const reopenTimeouts = [2000, 5000, 10000, 30000, 60000];

const changeRequestsDispatcher = (url, socketOptions, onSocketReconnect) => {
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
        reopenTimeoutHandler = setTimeout(() => {

                onSocketReconnect()
                openSocket().then(() => Turbo.session.connectStreamSource(socket)).catch(e => {

                })
            },
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
    return (id, action, target, content, params) => {
        if (!id) {
            throw 'changeRequest.id is required';
        }
        const changeRequest = {
            id: id,
            action: action,
            target: target,
            content_template: content,
            params: params
        }
        const send = () => socket.send(JSON.stringify(changeRequest));
        if (!socket || socket && socket.readyState !== WebSocket.OPEN) openSocket().then(send);
        else send();
    }
}