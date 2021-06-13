
var ipAddr = document
.getElementById("ip")
.innerHTML
if (ipAddr.includes(":")) {
document
    .getElementById("ip-report")
    .classList
    .add("border-green-700")
document
    .getElementById("ip-report-title")
    .classList
    .add("bg-green-700")
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
    .add("border-red-700")
document
    .getElementById("ip-report-title")
    .classList
    .add("bg-red-700")
document
    .getElementById("emoji")
    .innerHTML = "😿"
document
    .getElementById("message")
    .innerHTML = "You are using IPv4 to connect to this server."
};