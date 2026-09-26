import { createServer } from "node:http";

const upstream = process.env.MOCKPORT_UPSTREAM_URL;
let failNextCreate = false;

createServer(async (request, response) => {
  if (request.method === "POST" && request.url === "/_test/fail-next") {
    failNextCreate = true;
    response.writeHead(200, { "Content-Type": "application/json" });
    return response.end('{"armed":true}');
  }
  if (request.method === "POST" && request.url === "/v1/checkout/sessions") {
    await new Promise((resolve) => setTimeout(resolve, 50));
    if (failNextCreate) {
      failNextCreate = false;
      response.writeHead(503, { "Content-Type": "application/json" });
      return response.end('{"error":{"type":"api_error","message":"one-time test failure"}}');
    }
  }
  if (request.method === "GET" && request.url.startsWith("/v1/checkout/sessions/")) {
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  try {
    const headers = { ...request.headers };
    delete headers.host;
    const result = await fetch(new URL(request.url, upstream), {
      method: request.method,
      headers,
      body: request.method === "GET" || request.method === "HEAD" ? undefined : request,
      duplex: "half",
    });
    response.writeHead(result.status, Object.fromEntries(result.headers));
    response.end(Buffer.from(await result.arrayBuffer()));
  } catch (error) {
    response.writeHead(502, { "Content-Type": "application/json" });
    response.end(JSON.stringify({ error: String(error) }));
  }
}).listen(Number(process.env.PORT || 43103), "127.0.0.1");
