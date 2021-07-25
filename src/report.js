var ipAddr = document
    .getElementById("ip")
    .innerHTML
if (ipAddr.includes(":")) {
    document
        .getElementById("ip-report")
        .classList
        .add("border-blue")
    document
        .getElementById("ip-report-title")
        .classList
        .add("bg-blue")
    document
        .getElementById("emoji")
        .innerHTML = "😸"
    document
        .getElementById("message")
        .innerHTML = "You are using IPv6 to connect to this server!"
    confetti.start(1200, 50, 150)
} else {
    document
        .getElementById("ip-report")
        .classList
        .add("border-pink")
    document
        .getElementById("ip-report-title")
        .classList
        .add("bg-pink")
    document
        .getElementById("emoji")
        .innerHTML = "😿"
    document
        .getElementById("message")
        .innerHTML = "You are using IPv4 to connect to this server."
};