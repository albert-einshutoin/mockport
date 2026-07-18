import OpenAI from "openai";

const mockportURL = process.env.MOCKPORT_BASE_URL || "http://127.0.0.1:43101";
const client = new OpenAI({
  apiKey: "mockport_openai_key",
  baseURL: new URL("/openai/v1", mockportURL).toString(),
});

const models = await client.models.list();
const completion = await client.chat.completions.create({
  model: "gpt-mockport",
  messages: [{ role: "user", content: "Hello from the Mockport example" }],
});

console.log(JSON.stringify({ model: models.data[0].id, completion }, null, 2));
