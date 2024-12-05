const ipAddr = document.getElementById("ip").innerHTML;
const ipReport = document.getElementById("ip-report");
const ipReportTitle = document.getElementById("ip-report-title");
const emoji = document.getElementById("emoji");
const message = document.getElementById("message");

const ipConfig = {
  ipv6: {
    borderClass: ["border-purple", "text-white"],
    titleClass: ["bg-purple"],
    emoji: "😸",
    msg: "You are using IPv6 to connect to this server!",
    confetti: true
  },
  ipv4: {
    borderClass: ["border-pink", "text-white"],
    titleClass: ["bg-pink"],
    emoji: "😿",
    msg: "You are using IPv4 to connect to this server."
  },
  unknown: {
    borderClass: ["border-yellow"],
    titleClass: ["bg-yellow", "text-black"],
    emoji: "🤔",
    msg: "We could not reliably determine your IP.",
    clearWallOfText: true
  }
};

function setIPDisplay(type) {
  const config = ipConfig[type];

  ipReport.classList.add(...config.borderClass);
  ipReportTitle.classList.add(...config.titleClass);
  emoji.innerHTML = config.emoji;
  message.innerHTML = config.msg;

  if (config.confetti) {
    confetti.start(1200, 50, 150);
  }

  if (config.clearWallOfText) {
    document.getElementById("wall-of-text").innerHTML = "";
  }
}

if (ipAddr.includes(":")) {
  setIPDisplay("ipv6");
} else if (ipAddr.includes(".")) {
  setIPDisplay("ipv4");
} else {
  setIPDisplay("unknown");
}
