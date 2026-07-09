"use strict";

const fs = require("node:fs");
const path = require("node:path");
const OpenAI = require("openai");
const { toFile } = require("openai/uploads");

const streamCompletionFullText = "Mockport simulated streaming response.";

function openAISDKLabel() {
  const packagePath = path.join(path.dirname(require.resolve("openai")), "package.json");
  const packageData = JSON.parse(fs.readFileSync(packagePath, "utf8"));
  return `openai@${packageData.version}`;
}

async function runOpenAISmoke(options) {
  const client = new OpenAI({
    apiKey: "mockport_openai_key",
    baseURL: new URL("/openai/v1", options.baseURL).toString(),
    maxRetries: 0,
  });

  const models = await client.models.list();
  const chat = await client.chat.completions.create({
    model: "gpt-mockport",
    messages: [{ role: "user", content: "hello" }],
  });
  const stream = await client.chat.completions.create({
    model: "gpt-mockport",
    messages: [{ role: "user", content: "stream" }],
    stream: true,
  });
  let streamed = "";
  let streamChunkCount = 0;
  for await (const chunk of stream) {
    streamChunkCount++;
    streamed += chunk.choices?.[0]?.delta?.content || "";
  }
  const response = await client.responses.create({
    model: "gpt-mockport",
    input: "hello",
  });
  const retrievedResponse = await client.responses.retrieve(response.id);
  const embedding = await client.embeddings.create({
    model: "text-embedding-mockport",
    input: "hello",
  });
  const file = await client.files.create({
    purpose: "batch",
    file: await toFile(Buffer.from("{\"custom_id\":\"one\"}\n"), "mockport.jsonl"),
  });
  const batch = await client.batches.create({
    input_file_id: file.id,
    endpoint: "/v1/responses",
    completion_window: "24h",
  });

  assertEqual(models.data[0].id, "gpt-mockport", "model id");
  assertEqual(chat.object, "chat.completion", "chat object");
  assertEqual(retrievedResponse.id, response.id, "retrieved response id");
  assertEqual(embedding.object, "list", "embedding object");
  if (embedding.data[0].embedding.length < 3) {
    throw new Error(`embedding vector too small: ${embedding.data[0].embedding.length}`);
  }
  assertEqual(file.purpose, "batch", "file purpose");
  assertEqual(batch.input_file_id, file.id, "batch input file id");
  if (streamChunkCount < 3) {
    throw new Error(`stream chunk count too small: ${streamChunkCount}`);
  }
  if (streamed !== streamCompletionFullText) {
    throw new Error(`streamed content mismatch: ${streamed}`);
  }

  await assertChatCompletionStreamSSEContract(options.baseURL);

  return {
    provider: "openai",
    baseURL: options.baseURL,
    status: "sdk-ok",
    sdk: openAISDKLabel(),
    chatCompletion: chat.id,
    response: response.id,
    embedding: embedding.data[0].embedding.length,
    file: file.id,
    batch: batch.id,
  };
}

function assertEqual(got, want, label) {
  if (got !== want) {
    throw new Error(`${label}: got ${got}, want ${want}`);
  }
}

async function openAIRawPost(baseURL, endpoint, body) {
  const response = await fetch(new URL(`/openai/v1/${endpoint}`, baseURL), {
    method: "POST",
    headers: {
      "authorization": "Bearer mockport_openai_key",
      "content-type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const text = await response.text();
  return { response, text };
}

function parseSSEDataEvents(body) {
  const events = [];
  for (const block of body.split("\n\n")) {
    if (block === "") {
      continue;
    }
    const lines = block.split("\n");
    if (lines.length !== 1 || !lines[0].startsWith("data: ")) {
      throw new Error(`invalid SSE event framing: ${JSON.stringify(block)}`);
    }
    events.push(lines[0].slice("data: ".length));
  }
  return events;
}

async function assertChatCompletionStreamSSEContract(baseURL) {
  const { response, text } = await openAIRawPost(baseURL, "chat/completions", {
    model: "gpt-mockport",
    messages: [{ role: "user", content: "stream contract" }],
    stream: true,
  });
  if (!response.ok) {
    throw new Error(`chat stream status=${response.status} body=${text}`);
  }
  const contentType = response.headers.get("content-type") || "";
  if (!contentType.startsWith("text/event-stream")) {
    throw new Error(`chat stream content-type=${contentType}, want text/event-stream`);
  }

  const events = parseSSEDataEvents(text);
  if (events.length < 3) {
    throw new Error(`chat stream event count too small: ${events.length}`);
  }
  if (events[events.length - 1] !== "[DONE]") {
    throw new Error(`chat stream terminal event=${events[events.length - 1]}, want [DONE]`);
  }

  const dataEvents = events.slice(0, -1);
  let assembled = "";
  let contentDeltaCount = 0;
  for (const [index, raw] of dataEvents.entries()) {
    let chunk;
    try {
      chunk = JSON.parse(raw);
    } catch (error) {
      throw new Error(`chat stream chunk ${index} is not JSON: ${raw}`);
    }
    const deltaContent = chunk.choices?.[0]?.delta?.content;
    if (typeof deltaContent === "string" && deltaContent !== "") {
      contentDeltaCount += 1;
      assembled += deltaContent;
    }
  }
  if (contentDeltaCount < 1) {
    throw new Error(`chat stream missing content deltas: ${contentDeltaCount}`);
  }
  if (assembled !== streamCompletionFullText) {
    throw new Error(`chat stream assembled content=${assembled}, want ${streamCompletionFullText}`);
  }

  const responses = await openAIRawPost(baseURL, "responses", {
    model: "gpt-mockport",
    input: "stream contract",
    stream: true,
  });
  if (!responses.response.ok) {
    throw new Error(`responses stream status=${responses.response.status} body=${responses.text}`);
  }
  const responsesContentType = responses.response.headers.get("content-type") || "";
  if (!responsesContentType.startsWith("application/json")) {
    throw new Error(`responses stream content-type=${responsesContentType}, want application/json`);
  }
  if (responses.text.includes("data: [DONE]")) {
    throw new Error(`responses stream returned SSE terminal marker: ${responses.text}`);
  }
  let responsesBody;
  try {
    responsesBody = JSON.parse(responses.text);
  } catch (error) {
    throw new Error(`responses stream body is not JSON: ${responses.text}`);
  }
  if (responsesBody.object !== "response") {
    throw new Error(`responses stream object=${responsesBody.object}, want response`);
  }
}

module.exports = { runOpenAISmoke };
