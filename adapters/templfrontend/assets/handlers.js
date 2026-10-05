function search(el) {
    const filter = encodeURIComponent(el.querySelector("input.search-box").value);

    if (!!filter) {
        window.location.href = "/search/" + filter;
    } else {
        window.location.href = "/";
    }

    return false;
}

let hoverExampleId = "";
let hoverWordIds = [];

function getHoverableParent(el) {
    while (["A", "a", "SPAN", "span", "U", "u"].includes(el.nodeName)) {
        if (!!el.dataset.ids) {
            break;
        }

        el = el.parentNode;
    }

    return el;
}

function onHover(el) {
    el = getHoverableParent(el);

    const ids = JSON.parse(el.dataset.ids || "[]");
    const extraIds = [];
    const exampleNode = el.parentNode.parentNode;
    const exampleId = el.parentNode.dataset.id;

    if (!!hoverExampleId) {
        const prev = exampleNode.querySelectorAll("span, a");
        for (const el of prev) {
            el.classList.remove("hover");
        }
    }

    if (ids.length === 0) {
        return;
    }

    const current = exampleNode.querySelectorAll("span, a");
    for (const partEl of current) {
        if (!partEl.dataset.ids) {
            continue;
        }

        const ids2 = JSON.parse(partEl.dataset.ids);
        if (ids2.find(id2 => ids.find(id => id === id2))) {
            for (const id of ids2) {
                if (!extraIds.includes(id)) {
                    extraIds.push(id);
                }
            }

            partEl.classList.add("hover");
        }
    }

    for (const partEl of current) {
        if (!partEl.dataset.ids) {
            continue;
        }

        const ids2 = JSON.parse(partEl.dataset.ids);
        if (ids2.find(id2 => extraIds.find(id => id === id2))) {
            partEl.classList.add("hover");
        }
    }

    hoverExampleId = exampleId;
    hoverWordIds = ids;
}

function onHoverEnd(el) {
    el = getHoverableParent(el);

    const ids = JSON.parse(el.dataset.ids || "[]");
    const exampleId = el.parentNode.dataset.id;
    const exampleNode = el.parentNode.parentNode;

    if (hoverExampleId === exampleId && JSON.stringify(ids) === JSON.stringify(hoverWordIds)) {
        const prev = exampleNode.querySelectorAll("span, a");
        for (const el of prev) {
            el.classList.remove("hover");
        }
    }
}

// This is so that the textbox changes back if you go back after searching.
window.addEventListener("DOMContentLoaded", function() {
    // Fill the search box with the filter, so that going back will replace it.
    let filter = decodeURIComponent(window.location.pathname.split("/").pop())
    if (!!window.location.searchParams && window.location.searchParams.has("q")) {
        filter = decodeURIComponent(window.location.searchParams.get("q"));
    }
    const searchBox = document.querySelector("input.search-box");
    searchBox.value = filter;

    let exampleDivs = document.querySelectorAll(".example").values();
    const addListenersBatch = function() {
        console.time("addListenersBatch");

        const current = [];
        while (current.length < 100) {
            const v = exampleDivs.next();
            if (v.done) {
                break
            }
            current.push(v.value);
        }

        for (const el of current) {
            const buttonRow = el.querySelector("div.button-row")
            if (buttonRow != null) {
                const quoteDiscordButton = buttonRow.querySelector("button.quote-discord");
                if (quoteDiscordButton != null) {
                    quoteDiscordButton.onclick = generatePastableQuote.bind(el, el.id, filter, "discord")
                }
                const quoteForumButton = buttonRow.querySelector("button.quote-forum");
                if (quoteForumButton != null) {
                    quoteForumButton.onclick = generatePastableQuote.bind(el, el.id, filter, "bbcode")
                }
            }

            el.addEventListener('mouseover', function(event) {
                if (event.target) {
                    onHover(event.target);
                }
            })
            el.addEventListener('mouseout', function(event) {
                if (event.target) {
                    onHoverEnd(event.target);
                }
            })
        }

        if (current.length === 100) {
            requestAnimationFrame(addListenersBatch);
        }

        console.timeEnd("addListenersBatch");
    }
    setTimeout(addListenersBatch, 0);
});

const lastFormat = {}

function generatePastableQuote(exampleElementId, filter, format) {
    const [_, filterIndex, ...exampleIdParts] = exampleElementId.split("-");
    const exampleId = exampleIdParts.join("-");
    console.log(exampleElementId, exampleId, filter, filterIndex);
    let outputPre = this.querySelector("pre.copy-paste-text");
    if (outputPre == null) {
        outputPre = document.createElement("pre");
        outputPre.className = "copy-paste-text";
        this.append(outputPre)
    }

    if (lastFormat[exampleElementId] === format) {
        outputPre.remove();
        lastFormat[exampleElementId] = "";
    } else {
        lastFormat[exampleElementId] = format;
    }

    outputPre.textContent = "Loading...";

    fetch(`/api/examples/${exampleId}/discord-quote?filter=${encodeURIComponent(filter)}&format=${format}&filter_index=${filterIndex}`)
        .then(res => {
            return res.json();
        }).then(data => {
            outputPre.textContent = data.text;
        }).catch(err => {
            outputPre.textContent = "REQUEST FAILED:\n" + err.toString()
        })
}