CREATE TEMP FUNCTION directories(filename STRING)
RETURNS ARRAY<STRING>
LANGUAGE js
AS r"""
  if (filename == null)
    return ["Unknown File"]
  if (!filename.startsWith("//"))
    return ["Invalid path"]
  var dirs = [];
  var path = "//"
  var parts = filename.split("/")
  for (let i = 0; i < parts.length; i++) {
  if (parts[i] == "")
    continue
  path += parts[i]
  dirs.push(path)
  path += "/"
  }

  return dirs;
""";

MERGE INTO {project}.{dataset}.daily_file_metrics AS T
USING (
  WITH file_summaries AS (
    SELECT
      `date`,
      file_name,
      component,
      ARRAY_AGG(test_id) AS test_ids,
      repo AS repo,
      SUM(num_runs) AS num_runs,
      SUM(num_failures) AS num_failures,
      SUM(num_flake) AS num_flake,
      SUM(total_runtime) AS total_runtime,
      -- The average for the file is still the sum of the tests contained within
      SUM(avg_runtime) AS avg_runtime,
      SUM(p50_runtime) AS p50_runtime,
      SUM(p90_runtime) AS p90_runtime,
    FROM
      {project}.{dataset}.daily_test_metrics AS day_metrics
    WHERE DATE(`date`) BETWEEN @from_date AND @to_date
    GROUP BY
      file_name, `date`, component, repo
  ), dir_nodes AS (
    SELECT
      node_name,
      -- null node names means it isn't a proper directory path to have been
      -- expanded into directory nodes
      IFNULL(node_name = f.file_name, true) as is_file,
      f.*,
    FROM file_summaries AS f, UNNEST(directories(f.file_name)) AS node_name
  )
  -- Combine the file metrics into the directory metrics (treating files as
  -- a single file directory)
  SELECT
    date,
    repo,
    component,
    node_name,
    ANY_VALUE(n.is_file) AS is_file,
    SUM(num_runs) AS num_runs,
    SUM(num_failures) AS num_failures,
    SUM(num_flake) AS num_flake,
    SUM(total_runtime) AS total_runtime,
    SUM(avg_runtime) AS avg_runtime,
    SUM(p50_runtime) AS p50_runtime,
    SUM(p90_runtime) AS p90_runtime,
  FROM dir_nodes n
  GROUP BY `date`, component, node_name, repo
  ) AS S
ON
  T.date = S.date
  AND T.date BETWEEN @from_date AND @to_date
  AND T.node_name = S.node_name
  AND (T.component = S.component OR (T.component IS NULL AND S.component IS NULL))
  AND (T.repo = S.repo OR (T.repo IS NULL AND S.repo IS NULL))
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime
WHEN NOT MATCHED THEN
  INSERT (`date`, repo, component, node_name, is_file, num_runs, num_failures, num_flake, total_runtime, avg_runtime, p50_runtime, p90_runtime)
  VALUES (`date`, repo, component, node_name, is_file, num_runs, num_failures, num_flake, total_runtime, avg_runtime, p50_runtime, p90_runtime)
