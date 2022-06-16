#!/usr/bin/env python3
#
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import argparse
import io
import os
import pandas as pd
import sys
import tensorboard
import tensorflow as tf
import numpy as np
import time

# The following lines adjust the granularity of reporting.
pd.options.display.max_rows = 10
pd.options.display.float_format = '{:.1f}'.format
root_logdir = os.path.join(os.curdir, "tf_logs")


def get_run_logdir():
  run_id = time.strftime("run_%Y_%m_%d-%H_%M_%S")
  return os.path.join(root_logdir, run_id)


def predict(args):
  if args.file and args.input:
    raise Exception('Must provide either a file or raw input')

  # Create the model
  model = tf.keras.models.load_model(args.model)

  if args.file:
    raw_df = pd.read_csv(filepath_or_buffer=args.file)
  else:
    input_data = io.StringIO(args.input)
    raw_df = pd.read_csv(filepath_or_buffer=input_data)

  git_average, file_average = model.get_layer(
      'normalization').adapt_mean.value().numpy()
  test_df = extract_features(
      raw_df, git_average=git_average, file_average=file_average)

  features = {name: np.array(value) for name, value in test_df.items()}
  predictions = model.predict(features)

  plain_text = '\n'.join([str(prediction[0]) for prediction in predictions])
  if args.output:
    with open(args.output, 'w') as predictions_file:
      predictions_file.write(plain_text)


def train(args):
  raw_df = pd.read_csv(filepath_or_buffer=args.train_data)
  cleaned_df = extract_features(raw_df)

  # Print the first rows of the pandas DataFrame.
  print('Dataframe head:\n')
  print(cleaned_df.head())
  print('Dataframe description:\n')
  print(cleaned_df.describe())

  train_df = cleaned_df.sample(frac=0.9, random_state=200)
  validation = cleaned_df.drop(train_df.index)

  features = {name: np.array(value) for name, value in train_df.items()}
  label = np.array(features.pop("Failed"))

  # Create the model
  feature_columns = []
  feature_columns.append(tf.feature_column.numeric_column("GitDistance"))
  feature_columns.append(tf.feature_column.numeric_column("FileDistance"))
  feature_layer = tf.keras.layers.DenseFeatures(feature_columns)

  normalizer = tf.keras.layers.Normalization(axis=1)
  normalizer.adapt(cleaned_df.drop('Failed', axis=1))

  model = tf.keras.Sequential([
      feature_layer,
      normalizer,
      tf.keras.layers.Dense(units=1, activation='sigmoid'),
  ])

  model.compile(
      optimizer='adam',
      loss='binary_crossentropy',
      metrics=[tf.keras.metrics.AUC(num_thresholds=100, name='auc')])

  validation_features = {
      name: np.array(value) for name, value in validation.items()
  }
  validation_label = np.array(validation_features.pop("Failed"))

  tensorboard_cb = tf.keras.callbacks.TensorBoard(get_run_logdir())
  model.fit(
      x=features,
      y=label,
      epochs=10,
      validation_data=(validation_features, validation_label),
      callbacks=[tensorboard_cb])

  model.save(args.output)


def extract_features(raw_df,
                     include_label=True,
                     git_average=None,
                     file_average=None):
  training_df = pd.DataFrame()
  training_df['GitDistance'] = raw_df['GitDistance'].astype('float32')
  training_df['FileDistance'] = raw_df['FileDistance'].astype('float32')

  # Distances are sometimes missing and need to use the average when not present
  if not git_average:
    git_average = training_df['GitDistance'].mean()
  training_df['GitDistance'].fillna(value=git_average, inplace=True)
  if not file_average:
    file_average = training_df['FileDistance'].mean()

  training_df['FileDistance'].fillna(value=file_average, inplace=True)

  if 'Failed' in raw_df and include_label:
    training_df['Failed'] = raw_df['Failed'].astype('float32')
  return training_df


def main():
  arg_parser = argparse.ArgumentParser()
  arg_parser.usage = __doc__
  subparsers = arg_parser.add_subparsers()

  predict_parser = subparsers.add_parser('predict')
  predict_parser.add_argument('--file', type=str)
  predict_parser.add_argument('--input', type=str)
  predict_parser.add_argument('--model', required=True, type=str)
  predict_parser.add_argument('--output', required=False, type=str)
  predict_parser.set_defaults(func=predict)

  train_parser = subparsers.add_parser('train')
  train_parser.add_argument('--train-data', required=True, type=str)
  train_parser.add_argument('--split-train', required=False, type=bool)
  train_parser.add_argument('--test-data', required=False, type=str)
  train_parser.add_argument('--output', required=True, type=str)
  train_parser.set_defaults(func=train)

  args = arg_parser.parse_args()

  args.func(args)

  return 0


if __name__ == '__main__':
  sys.exit(main())
