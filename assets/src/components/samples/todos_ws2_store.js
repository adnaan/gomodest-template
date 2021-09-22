import websocketStore from "svelte-websocket-store";
let url = "ws://localhost:3000/samples/ws2"
if (process.env.ENV === "production"){
    url = `wss://${process.env.HOST}/samples/ws2`
}
export const todos = websocketStore(url, []);