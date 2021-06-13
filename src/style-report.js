var ipAddr = document.getElementById("ip").innerHTML
if (ipAddr.includes(":")) {
  document.getElementById("ip").classList.add("happy-green", "ip6-addr-mobile")
  document.getElementById("emoji").innerHTML = "😸"
  document.getElementById("message").innerHTML =
    "You are using IPv6 to connect to this server!"
  confetti.start(1200, 50, 150)
} else {
  document.getElementById("ip").classList.add("sad-red")
  document.getElementById("emoji").innerHTML = "😿"
  document.getElementById("message").innerHTML =
    "You are using IPv4 to connect to this server."
}
