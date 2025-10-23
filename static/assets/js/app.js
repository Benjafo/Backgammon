async function login() {
    try {
        await fetch(`/login`);
    } catch (e) {
        console.error(`Login failed`, e);
    }
}

async function fetchTurn() {
    const res = await fetch(`/turn`),
          data = await res.json();
    document.getElementById(`turn`).textContent = data.currentTurn;
}

async function nextTurn() {
    const res = await fetch(`/next`),
          data = await res.json();
    document.getElementById(`turn`).textContent = data.nextTurn;
}

document.addEventListener(`DOMContentLoaded`, () => {
    document.getElementById(`next-turn-btn`).addEventListener(`click`, nextTurn);
    (async () => {
        await login();
        await fetchTurn();
    })();
});
