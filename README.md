# alfred-bigquery-shortcuts
Alfred Shortcuts for Google Cloud BigQuery

## What's this thing
For you, dear person who works on Google BigQuery every day and is always lost in many clicks in the Web UI
around different projects and datasets, here's an [Alfred](https://www.alfredapp.com) Workflow to make your life a bit easier! 🤩

## Installation

1. Have Alfred with PowerPack installed
2. Download the latest release of `alfred-bigquery-shortcuts.alfredworkflow` from the [Releases](https://github.com/ralbertazzi/alfred-bigquery-shortcuts/releases) page
3. Double click
4. Enjoy 🍒

## Before Playing

Before enjoying this workflow you need to run the `bq-refresh` command.
This command will fetch the list of Google Cloud projects, BigQuery datasets and tables for which you have authenticated and
will cache this in your Mac. Depending on how many projects, datasets and tables you have, this command may take from a couple
of seconds to one minute.

_NOTE 1_: to make this workflow access the Google Cloud APIs, you need to have the `gcloud` CLI installed and be
authenticated through `gcloud auth application-default login`

_NOTE 2_: the workflow defines a maximum number of tables per dataset that will be fetched. This is done on purpose to avoid
spending a lot of time fetching information from date-sharded tables (where each shard is a separate table). If you really
wish to fetch even more tables, increase the value of `max_tables_per_dataset` from the "variables" section of the workflow.

_NOTE 3_: you need to run this command every time new datasets and tables are added to your projects

## Usage

There are three simple and similar commands:
* `bq`: select project -> select dataset -> select table -> open the BigQuery UI on the selected table
* `bqd`: select project -> select dataset -> open the BigQUery UI on the selected dataset
* `bqp`: select project -> open the BigQuery UI on the selected project

## Contribution

This is my first attempy of writing Go code and I think the code quality is really bad 🤢

Feel free to suggest improvements / new features!
