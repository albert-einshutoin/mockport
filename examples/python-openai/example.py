import json
import os

from openai import OpenAI


mockport_url = os.environ.get("MOCKPORT_BASE_URL", "http://127.0.0.1:43101")
client = OpenAI(
    api_key="mockport_openai_key",
    base_url=f"{mockport_url.rstrip('/')}/openai/v1",
)

models = client.models.list()
completion = client.chat.completions.create(
    model="gpt-mockport",
    messages=[{"role": "user", "content": "Hello from the Mockport Python example"}],
)

print(
    json.dumps(
        {
            "model": models.data[0].id,
            "completion": completion.model_dump(),
        },
        indent=2,
    )
)
