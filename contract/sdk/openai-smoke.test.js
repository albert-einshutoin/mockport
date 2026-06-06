"use strict";

const OpenAI = require("openai");
const { toFile } = require("openai/uploads");

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
  for await (const chunk of stream) {
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
  if (!streamed.includes("Mockport response")) {
    throw new Error(`streamed content missing Mockport response: ${streamed}`);
  }

  return {
    provider: "openai",
    baseURL: options.baseURL,
    status: "sdk-ok",
    sdk: "openai@6.39.1",
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

module.exports = { runOpenAISmoke };
