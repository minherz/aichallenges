# Tuning a Gen AI model

This challenge uses Vertex AI [Gemini supervised tuning](https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini-supervised-tuning).
It is simple and straightforward on the positive side.
On the negative side it is limited to only gemini models and it constraints the training dataset format to follow the [predefined format](https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini-supervised-tuning-prepare#dataset-format).
The documentation has two different pages about the format:

* [Prepare your data](https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini-supervised-tuning-prepare#dataset-format) describes the format for mulitple modalities
* [Supported modalities / text tuning](https://cloud.google.com/vertex-ai/generative-ai/docs/models/tune_gemini/text_tune) describes the same but for text modality only

Majority of [datasets on hugging faces](https://huggingface.co/datasets) either come in the different (non JSONL) file formats or do not follow the same JSON schema.
So, the task of tuning the model is converted to the two tasks:

1. Convert the dataset to the format supported by Vertex AI
1. Use Vertex AI supervised tuning to train the model

I ended up with writing a simple CLI to reformat the dataset. It converts JSONL file in the "known" format to the expected format.

> [!WARNING]
> Validation dataset should be limited to 256 entries maximum.
