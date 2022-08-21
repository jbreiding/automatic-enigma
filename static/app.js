function makeRow(container, date, games, teams) {
    games.sort((a, b) => new Date(a.Time) - new Date(b.Time))

    var header = document.createElement("h2")
    var d = document.createElement("strong")
    const options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' };
    d.append(date.toLocaleDateString("en-US", options))
    header.append(d)

    container.append(header)

    var row = document.createElement("div")
    row.className = "grid"

    var count = 0
    games.forEach(game => {
        makeCard(row, game)
        count++
    })

    if (count < teams) {
        while (count < teams) {
            blank = document.createElement("article")
            blank.style.visibility = "hidden"
            row.append(blank)
            count++
        }
    }

    container.append(row)
}

function makeCard(container, game) {
    var card = document.createElement("article")
    var cardHeader = document.createElement("header")
    cardHeader.append(`${game.Team} (${game.Home ? "HOME" : "AWAY"})`)
    cardHeader.append(document.createElement("br"))
    cardHeader.append(`${game.Home ? "vs" : "at"}`)
    cardHeader.append(document.createElement("br"))
    cardHeader.append(`${game.Opponent}`)

    card.append(cardHeader)

    card.append(new Date(game.Time).toLocaleTimeString())

    var link = document.createElement("a")
    link.setAttribute("href", `https://home.sportsaffinity.com/m/tour/venueinfo.aspx?venueguid=${game.Venue}`)
    link.setAttribute("role", "button")
    link.append(`@ ${game.Location}`)
    var small = document.createElement("small")
    small.append(link)

    var td = document.createElement("footer")
    td.append(small)
    card.append(td)
    container.append(card)
}

function load(count) {
    var games = document.getElementById("table-games");
    games.innerHTML = '';
    fetch("teams.json")
        .then(response => {
            return response.json();
        })
        .then(jsondata => {
            Object.entries(jsondata).sort((a, b) => new Date(a[0]) - new Date(b[0])).forEach(event => {
                makeRow(games, new Date(event[0]), event[1], count)
            })
            return jsondata
        });
}