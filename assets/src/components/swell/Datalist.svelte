<script>
    import websocketStore from "svelte-websocket-store";
    import {onDestroy, onMount} from "svelte";

    export let resource;
    export let url;
    export let socketOptions = [];

    const opCreate = "create";
    const opDelete = "delete";
    const opUpdate = "update";
    const opList = "list"
    let methodID = 0;
    let unsubscribe;
    let store = websocketStore(url, socketOptions);
    let items = [];
    let operations = new Map();

    const call = (resource, method, params) => {
        methodID += 1
        operations.set(methodID, method);
        return {
            jsonrpc: "2.0",
            method: resource ? `${resource}/${method}` : method,
            id: methodID,
            params: params
        }
    }

    const ref = {
        create: (item) => $store = call(resource, opCreate, item),
        delete: (item) => $store = call(resource, opDelete, item),
        update: (item) => $store = call(resource, opUpdate, item),
    }

    // Props changed
    $: {
        if (unsubscribe) {
            unsubscribe();
            store = websocketStore(url, []);
        }
        unsubscribe = store.subscribe(data => {
            console.log(data)
            if (data.result) {
                const op = operations.get(data.id)
                operations.delete(data.id)
                console.log("op => ", op)
                switch (op) {
                    case opList:
                        if (data.result.length > 0) {
                            items = data.result;
                        }
                        break;
                    case opCreate:
                        items = [...items, data.result]
                        break;
                    case opUpdate:
                        items = items.map(item => (item.id === data.result.id) ? data.result : item)
                        break;
                    case opDelete:
                        items = items.filter(item => item.id !== data.result.id)
                        break;
                    default:
                        console.log(`orphan response: ${data.id}`)
                }

            }
        });
    }
    onMount(() => $store = call(resource, opList))
    onDestroy(() => unsubscribe());
</script>

<div>
    <slot items={items} ref={ref}/>
    {#if store.loading}
        <slot name="loading"/>
    {:else}
        <slot name="fallback"/>
    {/if}
</div>