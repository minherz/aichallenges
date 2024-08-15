function extractIdsFromString(message) {
    const idPattern = /\[([a-zA-Z0-9-]+)\]/g;
    const matches = message.matchAll(idPattern);
    const ids = [];
    for (const match of matches) {
        ids.push(match[1]);
    }

    return ids;
}

function createUserMessage(message) {
    const usermessage = document.createElement("p");
    const userMessageSpan = document.createElement("span");
    userMessageSpan.innerText = message;
    userMessageSpan.classList.add("user-message-text");
    usermessage.classList.add("user-message");
    usermessage.appendChild(userMessageSpan);
    return usermessage;
}

var sessionId = ""
const botmessages = document.getElementById("bot-messages");
const botbutton = document.getElementById("bot-input-button");
const botinput = document.getElementById("bot-input-text");

async function main() {
    botbutton.addEventListener("click", handleButtonClick);
    botinput.addEventListener("keypress", (event) => {
        if (event.key === "Enter") {
            botbutton.click();
        }
    });
}

async function handleButtonClick() {
    console.log("bot button clicked");
    if (!botinput.value || !botinput.value.trim) {
        return;
    }

    const message = botinput.value;
    console.log("message: " + message);
    const usermessage = createUserMessage(message);
    botmessages.appendChild(usermessage);
    botmessages.scrollTo(0, botmessages.scrollHeight);
    botinput.value = "";

    // Disable send button and input field
    botbutton.disabled = true;
    botinput.disabled = true;
    console.log("bot is typing");

    // Construct and render placeholder bot message
    const botmessage = document.createElement("p");
    botmessage.classList.add("bot-message-loading");
    const botmessagespan = document.createElement("span");
    botmessagespan.innerText = ""
    botmessagespan.classList.add("bot-message-loading");
    botmessage.classList.add("bot-message");
    botmessage.appendChild(botmessagespan);
    botmessages.appendChild(botmessage);
    botmessages.scrollTo(0, botmessages.scrollHeight);

    // Request a response from the Shopping Assistant
    const response = await fetch("ask", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            message: message,
            session: sessionId,
        }),
    });
    const responseJson = await response.json();
    console.log(responseJson);
    // refresh session Id
    sessionId = response.sessionId

    // Replace the placeholder bot message text with the real response
    // Making sure to remove any lists or product IDs from that message
    botmessagespan.innerText = responseJson.message;
    botmessage.classList.remove("bot-message-loading");
    botmessages.scrollTo(0, botmessages.scrollHeight);

    // Re-enable button and input field
    botbutton.disabled = false;
    botinput.disabled = false;
    botinput.focus();
}

main();
