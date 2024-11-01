import pandas as pd
import json, os, sys
from sklearn.model_selection import train_test_split


def convert(df: pd.DataFrame, output_path: str) -> None:
    # remove unused fields based on prior knowledge about the dataset
    df = df[['instruction', 'response']]

    with open(output_path, "w") as f:
        for _, row in df.iterrows():
            formatted_data = {
                "contents": [
                    {
                        'role': 'user',
                        'parts': [
                            {
                                'text': row['instruction']
                            },
                        ]
                    },
                    {
                        'role': 'model',
                        'parts': [
                            {
                                'text': row['response']
                            }
                        ]
                    }
                ]
            }
            f.write(json.dumps(formatted_data) + "\n")


def main(dataset_path: str) -> None:
    basename = os.path.basename(dataset_path)
    if basename.lower().endswith('.jsonl'):
        df = pd.read_json(dataset_path, lines=True)
    elif basename.lower().endswith('.json'):
        df = pd.read_json(dataset_path, lines=False)
    elif basename.lower().endswith('.csv'):
        df = pd.read_csv(dataset_path)
    else:
        print('The format of the dataset file is not supported')
        exit(1)

    # debug printing
    print('dataset is loaded:')
    print(df.head(2))
    print('...')

    # maximum allowed number of validation samples is 256
    train_df, validation_df = train_test_split(df, test_size=256, random_state=41)
    filename = os.path.splitext(dataset_path)[0]
    convert(train_df, filename + '_train.jsonl')
    convert(validation_df, filename + '_verify.jsonl')

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print('Pass the path to the original YAML file as the argument.')
        exit(1)
    main(sys.argv[1])