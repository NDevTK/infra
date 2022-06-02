CREATE OR REPLACE VIEW `APP_ID.PROJECT_NAME.step_status_transitions`
AS
/*
Step status transition table.
This view represents status transitions for build steps over time.
Each row represents a build where the step produced a different status
output that it did in the previous run (on that builder etc).
*/
WITH
  step_lag AS (
  SELECT
    b.end_time AS end_time,
    project,
    bucket,
    builder,
    b.buildergroup as buildergroup,
    b.id,
    b.number,
    b.critical as critical,
    b.status as status,
    output_commit,
    input_commit,
    step.name AS step_name,
    step.status AS step_status,
    LAG(step.status) OVER (PARTITION BY project, bucket, builder, b.output_commit.host, b.output_commit.project, b.output_commit.ref, step.name ORDER BY b.output_commit.position, b.id desc, b.number) AS previous_step_status,
    LAG(b.output_commit) OVER (PARTITION BY project, bucket, builder, b.output_commit.host, b.output_commit.project, b.output_commit.ref, step.name ORDER BY b.output_commit.position, b.id desc, b.number) AS previous_output_commit,
     LAG(b.input_commit) OVER (PARTITION BY project, bucket, builder, b.input_commit.host, b.input_commit.project, b.input_commit.ref, step.name ORDER BY b.input_commit.position, b.id desc, b.number) AS previous_input_commit,
    LAG(b.id) OVER (PARTITION BY project, bucket, builder, b.output_commit.host, b.output_commit.project, b.output_commit.ref, step.name ORDER BY b.output_commit.position, b.number) AS previous_id
    FROM
    `sheriff-o-matic.materialized.buildbucket_completed_builds_prod` AS b cross join
    UNNEST(steps) AS step
  WHERE
    PROJECT_FILTER_CONDITIONS
)
SELECT
  end_time,
  project,
  bucket,
  builder,
  buildergroup,
  number,
  id,
  critical,
  status,
  output_commit,
  input_commit,
  step_name,
  step_status,
  previous_output_commit,
  previous_input_commit,
  previous_id,
  previous_step_status
FROM
  step_lag s
WHERE
  s.previous_step_status != s.step_status
