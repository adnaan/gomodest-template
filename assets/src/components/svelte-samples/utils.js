let todosURL = "ws://localhost:3000/samples/ws/todos"
if (process.env.ENV === "production") {
    todosURL = `wss://${process.env.HOST}/samples/ws/todos`
}

export {
    todosURL
}

