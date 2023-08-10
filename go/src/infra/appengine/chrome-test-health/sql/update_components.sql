CREATE OR REPLACE TABLE {project}.{dataset}.components AS (
  SELECT DISTINCT component
  FROM {project}.{dataset}.raw_metrics
  WHERE `date` > DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
  ORDER BY component
)