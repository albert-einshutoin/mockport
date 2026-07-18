# OpenAI Python SDK

This runnable example points the official OpenAI Python SDK at a local Mockport
process and uses a fake key only.

Start Mockport from the repository root:

```bash
go run ./cmd/mockport run --config examples/openai-chat/mockport.yml
```

Then run the example in a virtual environment:

```bash
cd examples/python-openai
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -r requirements.txt
python example.py
```
