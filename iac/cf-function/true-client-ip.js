async function handler(event) {
  const request = event.request;
  const clientIP = event.viewer.ip;
  const uri = event.request.uri;
  const userAgent = request.headers["user-agent"].value.toLowerCase();
  const isCliTool = /(^curl|^wget|^httpie)/.test(userAgent);

  const ALLOWED_PATHS = ["/", "/ip"];

  // Restrict access to allowed paths
  if (!ALLOWED_PATHS.includes(uri)) {
    return {
      statusCode: 403,
      body: "Access Forbidden",
      headers: {
        "content-type": { value: "text/plain" },
        "cache-control": { value: "no-store" },
      },
    };
  }
  request.headers["true-client-ip"] = {
    value: clientIP,
  };

  // Handle IP request for CLI tools or explicit /ip path
  if (isCliTool || uri === "/ip") {
    return {
      statusCode: 200,
      body: clientIP,
      headers: {
        "content-type": { value: "text/plain" },
        "cache-control": { value: "no-store" },
      },
    };
  }

  return request;
}
