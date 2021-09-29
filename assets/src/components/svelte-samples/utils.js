let todosURL = "ws://localhost:3000/samples/ws/todos"
if (process.env.ENV === "production") {
    todosURL = `wss://${process.env.HOST}/samples/ws/todos`
}

const todosConn = {
    url: todosURL,
    socketOptions: []
};

const todosChangeEventHandlers = {
    "error": (items, result) => console.error(result),
    "todos/list": (items, result) => {
        items = [...items, ...result];
        return items.sort(sortTodos)
    },
    "todos/insert": (items, result) => {
        items = [...items, result];
        return items.sort(sortTodos)
    },
    "todos/update": (items, result) => {
        items = items.map(item => (item.id === result.id) ? result : item);
        return items.sort(sortTodos)
    },
    "todos/delete": (items, result) => {
        items = items.filter(item => item.id !== result.id);
        return items.sort(sortTodos)
    },
}

const todoChangeEventHandlers = {
    "error": (item, result) => console.error(result),
    "todos/get": (item, result) => result,
    "todos/insert": (item, result) => window.location.href = "/samples/svelte_ws2_todos_multi",
    "todos/update": (item, result) => {
        return {...item, ...result}
    },
    "todos/delete": (item, result) => window.location.href = "/samples/svelte_ws2_todos_multi",
}

const sortTodos = (a, b) => {
    return new Date(b.created_at) - new Date(a.created_at)
}

export {
    todosURL,
    todosConn,
    todosChangeEventHandlers,
    todoChangeEventHandlers
}

