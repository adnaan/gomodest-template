import app from "./App.svelte"
import todos from "./svelte-samples/fetch/Todos.svelte"
import todows from "./svelte-samples/jsonrpc-websocket/TodosWs.svelte"
import todows2 from "./svelte-samples/jsonrpc2-websocket-spa/TodosWs2.svelte"
import todows2multilist from "./svelte-samples/jsonrpc2-websocket-mpa/List.svelte"
import todows2multiview from "./svelte-samples/jsonrpc2-websocket-mpa/View.svelte"
import todows2multinew from "./svelte-samples/jsonrpc2-websocket-mpa/New.svelte"

// export other components here.
export default {
    app: app,
    todos: todos,
    todows: todows,
    todows2: todows2,
    todows2multilist:todows2multilist,
    todows2multiview: todows2multiview,
    todows2multinew: todows2multinew
}