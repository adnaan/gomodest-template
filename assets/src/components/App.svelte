<script>
export let title; // this prop hydrated from the server
console.log(title)
</script>

<main class="container is-fluid">
    <div class="columns is-centered is-vcentered is-mobile py-5">
        <div class="column is-narrow" style="width: 70%">
            <h1 class="has-text-centered title">{title}</h1>
        </div>
    </div>
</main>