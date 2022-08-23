#!/usr/bin/env python3
#
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# [VPYTHON:BEGIN]
# wheel: <
#   name: "infra/python/wheels/tensorflow/${vpython_platform}"
#   version: "version:2.7.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/keras-py3"
#   version: "version:2.7.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/h5py/${vpython_platform}"
#   version: "version:3.6.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/absl-py-py3"
#   version: "version:0.11.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/tensorboard-py3"
#   version: "version:2.6.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/werkzeug-py3"
#   version: "version:2.0.1"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/google-auth-oauthlib-py3"
#   version: "version:0.4.5"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/requests-oauthlib-py3"
#   version: "version:1.3.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/oauthlib-py3"
#   version: "version:3.1.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/tensorboard-data-server-py3"
#   version: "version:0.6.1"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/tensorboard-plugin-wit-py3"
#   version: "version:1.8.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/markdown-py3"
#   version: "version:3.0.1"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/tensorflow-io-gcs-filesystem/${vpython_platform}"
#   version: "version:0.23.1"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/tensorflow-estimator-py3"
#   version: "version:2.7.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/astunparse-py3"
#   version: "version:1.6.3"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/termcolor-py2_py3"
#   version: "version:1.1.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/google-pasta-py3"
#   version: "version:0.2.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/gast-py3"
#   version: "version:0.4.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/keras-preprocessing-py2_py3"
#   version: "version:1.1.2"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/libclang/${vpython_platform}"
#   version: "version:12.0.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/flatbuffers-py2_py3"
#   version: "version:1.12"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/opt-einsum-py3"
#   version: "version:3.3.0"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/wrapt/${vpython_platform}"
#   version: "version:1.13.3"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/six-py2_py3"
#   version: "version:1.15.0"
# >
# wheel: <
#   name: "infra/python/wheels/google-auth-py2_py3"
#   version: "version:1.25.0"
# >
# wheel: <
#   name: "infra/python/wheels/pyasn1-py2_py3"
#   version: "version:0.4.5"
# >
# wheel: <
#   name: "infra/python/wheels/pyasn1_modules-py2_py3"
#   version: "version:0.2.4"
# >
# wheel: <
#   name: "infra/python/wheels/rsa-py2_py3"
#   version: "version:3.4.2"
# >
# wheel: <
#   name: "infra/python/wheels/cachetools-py2_py3"
#   version: "version:2.0.1"
# >
# wheel: <
#   name: "infra/python/wheels/numpy/${vpython_platform}"
#   version: "version:1.2x.supported.1"
# >
# wheel: <
#   name: "infra/python/wheels/requests-py2_py3"
#   version: "version:2.26.0"
# >
# wheel: <
#   name: "infra/python/wheels/certifi-py2_py3"
#   version: "version:2020.11.8"
# >
# wheel: <
#   name: "infra/python/wheels/idna-py2_py3"
#   version: "version:2.8"
# >
# wheel: <
#   name: "infra/python/wheels/urllib3-py2_py3"
#   version: "version:1.24.3"
# >
# wheel: <
#   name: "infra/python/wheels/charset_normalizer-py3"
#   version: "version:2.0.4"
# >
# wheel: <
#   name: "infra/python/wheels/grpcio/${vpython_platform}"
#   version: "version:1.44.0"
# >
# wheel: <
#   name: "infra/python/wheels/protobuf-py3"
#   version: "version:3.20.0"
# >
# wheel: <
#   name: "infra/python/wheels/typing-extensions-py3"
#   version: "version:4.0.1"
#   match_tag: <
#     platform: "manylinux1_x86_64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/pandas/${vpython_platform}"
#   version: "version:1.3.2.chromium.1"
#   not_match_tag: <
#     platform: "linux_aarch64"
#   >
# >
# wheel: <
#   name: "infra/python/wheels/pytz-py2_py3"
#   version: "version:2018.4"
# >
# wheel: <
#   name: "infra/python/wheels/python-dateutil-py2_py3"
#   version: "version:2.7.3"
# >
# [VPYTHON:END]

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
