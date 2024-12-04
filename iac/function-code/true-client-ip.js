async function handler(event) {
  var request = event.request;
  var clientIP = event.viewer.ip;
  var userAgent = request.headers["user-agent"].value.toLowerCase();

  if (
    /(^curl|^wget|^httpie)/.test(userAgent) ||
    request.uri.startsWith("/ip")
  ) {
    var response = {
      body: clientIP,
      statusCode: 200,
      headers: {
        "content-type": {
          value: "text/plain; charset=UTF-8",
        },
        "cache-control": {
          value: "no-store",
        },
      },
    };
    return response;
  }

  request.headers["true-client-ip"] = {
    value: clientIP,
  };

  return request;
}
