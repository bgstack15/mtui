import events, { EVENT_STATS } from "../events.js";

const store = Vue.reactive({
    max_lag: null,
    uptime: null,
    time_of_day: null,
    player_count: null
});

events.on(EVENT_STATS, function(stats) {
    Object
    .keys(stats)
    .forEach(k => {
        store[k] = stats[k];
    });
});

export default store;