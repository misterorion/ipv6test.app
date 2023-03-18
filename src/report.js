var ipAddr = document.getElementById("ip").innerHTML;

switch (ipAddr) {
  case ipAddr.includes(":"):
    document
      .getElementById("ip-report")
      .classList.add("border-purple", "text-white");
    document.getElementById("ip-report-title").classList.add("bg-purple");
    document.getElementById("emoji").innerHTML = "😸";
    document.getElementById("message").innerHTML =
      "You are using IPv6 to connect to this server!";
    confetti.start(1200, 50, 150);
    break;
  case ipAddr.includes("."):
    document
      .getElementById("ip-report")
      .classList.add("border-pink", "text-white");
    document.getElementById("ip-report-title").classList.add("bg-pink");
    document.getElementById("emoji").innerHTML = "😿";
    document.getElementById("message").innerHTML =
      "You are using IPv4 to connect to this server.";
    break;
  default:
    document.getElementById("ip-report").classList.add("border-yellow");
    document
      .getElementById("ip-report-title")
      .classList.add("bg-yellow", "text-black");
    document.getElementById("emoji").innerHTML = "🤔";
    document.getElementById("message").innerHTML =
      "We could not reliably determine your IP.";
    document.getElementById("wall-of-text").innerHTML = "";
}
