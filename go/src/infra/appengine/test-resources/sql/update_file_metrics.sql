CREATE TEMP FUNCTION directories(filename STRING)
RETURNS ARRAY<STRING>
LANGUAGE js
AS r"""
  if (filename == null || !filename.startsWith("//"))
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

MERGE INTO %s.test_results.file_metrics AS T
USING (
  WITH file_summaries AS (
    SELECT
      date,
      file_name,
      component,
      ARRAY_AGG(test_id) AS test_ids,
      ANY_VALUE(repo) AS repo,
      ANY_VALUE(project) AS project,
      SUM(num_runs) AS num_runs,
      SUM(num_failures) AS num_failures,
      SUM(num_flake) AS num_flake,
      AVG(avg_runtime)	AS avg_runtime,
      SUM(total_runtime) AS total_runtime,
    FROM
      %s.test_results.test_metrics AS day_metrics
    WHERE DATE(date) BETWEEN @from_date AND @to_date
    GROUP BY
      file_name, date, component
  ), dir_nodes AS (
    SELECT
      node_name,
      node_name = f.file_name as is_file,
      f.*,
    FROM file_summaries AS f, UNNEST(directories(f.file_name)) AS node_name
  )

  SELECT
    date,
    component,
    node_name,
    ARRAY_REVERSE(SPLIT(node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
    ANY_VALUE(n.is_file) AS is_file,
    SUM(num_runs) AS num_runs,
    SUM(num_failures) AS num_failures,
    SUM(num_flake) AS num_flake,
    SUM(avg_runtime)	AS avg_runtime,
    SUM(total_runtime) AS total_runtime,
  FROM dir_nodes n
  GROUP BY date, component, node_name
  ) AS S
ON
  T.date = S.date
  AND T.component = S.component
  AND T.node_name = S.node_name
WHEN MATCHED THEN
  UPDATE SET
    num_runs = S.num_runs,
    num_failures = S.num_failures,
    num_flake = S.num_flake,
    avg_runtime = S.avg_runtime,
    total_runtime = S.total_runtime
WHEN NOT MATCHED THEN
  INSERT ROW
