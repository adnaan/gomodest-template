let todosURL = "ws://localhost:3000/samples/ws/todos"
if (process.env.ENV === "production") {
    todosURL = `wss://${process.env.HOST}/samples/ws/todos`
}

const todosReducers = {
    "error": (items, result) => console.error(result),
    "todos/list": (items, result) =>  [...items, ...result],
    "todos/insert": (items, result) => [...items, result],
    "todos/update": (items, result) => items.map(item => (item.id === result.id) ? result : item),
    "todos/delete": (items, result) => items.filter(item => item.id !== result.id),
}

const todoReducer = {
    "error": (item, result) => console.error(result),
    "todos/get": (item, result) => result,
    "todos/insert": (item, result) => window.location.href = "/samples/svelte_ws2_todos_multi",
    "todos/update": (item, result) => {
        return {...item, ...result}
    },
    "todos/delete": (item, result) => window.location.href = "/samples/svelte_ws2_todos_multi",
}


export {
    todosURL,
    todosReducers,
    todoReducer
}

