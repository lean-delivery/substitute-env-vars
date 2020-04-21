# Introduction to SEV
How many times have you been facing a situation when you have some dockerized application which stores its configuration on the filesystem?

What's a problem you may ask.
In a usual Docker world we would prefer to override or even intialize configuration parameters over environment variables.

Why is that ?
By design all Docker containers are stateless. 

That means, by default, working with stateful techniques in Docker (filesystem is one of those) increases complexity and operational costs.


# Real-life example
## Introductory
Let's assume we have a single-page-application (ReactJS). That application consists of some usual `static content`:
* HTML
* CSS
* JavaScript
* Assets (images, icons, etc.)

We have `3` environments:
* development
* qa
* production

And using `Adobe Analytics` for collecting data of user activities.


**Important Note**
We must segregate user activities data between environments, meaning each environment has its unique `Adobe Analytics URL` and `Google Analytics URL` defined in `index.html` file.
```html=
...
<script
  src="//assets.adobedtm.com/some/path/analytics-for-environment.min.js"
  async
></script>
<script
  src="test.url.com/analytics.js?utm_source=newsletter&utm_medium=banner&utm_campaign=spring_sale"
  async
></script>
...
```

## Dockerizing
For usual web applications we normally use `nginx` Docker image and populate it with `static content`.
```Dockerfile=
FROM nginx:alpine

# copy nginx configuration
COPY nginx/default.conf /etc/nginx/conf.d/

# copy static content
COPY dist/ /html

# run nginx
CMD ["/bin/sh", "-c", "nginx -g 'daemon off;'"]
```

## Problem statement
How would you deploy that `Docker image` and adjust `Adobe Analytics URL` and `Google Analytics URL` according to the specifc `environment` ?

**Option 1**
Compile `static content` per `environment`.
\- very time consuming and inefficient

**Option 2**
Set `environment variables` and utilize `bash`.
\- tricky, complex and messy

**Option 3**
Many `other options` are available, but with `operational and complexity overheads`. 

# Solution
This is where the `sev` comes handy.

## Introduction
`sev` is a CLI utility written in Golang.

## Usage example
### Let's change our `index.html`:
```html=
...
<script
  src="_{ADOBE_ANALYTICS_URL}_"
  async
></script>
<script
  src="_{GOOGLE_ANALYTICS_URL}_"
  async
></script>
...
```

### Good time to put `sev` in action
```sh
# First, let's define and intialize environment variables:
export ADOBE_ANALYTICS_URL=//assets.adobedtm.com/sc99ai/a9da01/launch-dif9-staging.min.js
export GOOGLE_ANALYTICS_URL=//test.url.com/?utm_source=newsletter&utm_medium=banner&utm_campaign=spring_sale

export VAR_NAMES_STORAGE=ADOBE_ANALYTICS_URL,GOOGLE_ANALYTICS_URL

# execute sev
sev /html/index.html
```

### Sample `sev`  output
```sh
##### START SEV ANNOUNCEMENT ##### 
 
## Target environment variables 
  ADOBE_ANALYTICS_URL=//assets.adobedtm.com/sc99ai/a9da01/launch-dif9-staging.min.js
  GOOGLE_ANALYTICS_URL=//test.url.com/?utm_source=newsletter&utm_medium=banner&utm_campaign=spring_sale
 
 
## Sets of values 
N/A 
 
## Mode 
  Pulling values for variables from environment 
 
###### END SEV ANNOUNCEMENT ###### 
 
2020/04/20 13:01:51 INFO: sev succeeded: /html/index.html 
```

### Checking `index.html` after `sev` processing
```sh=
...
<script
  src="//assets.adobedtm.com/sc99ai/a9da01/launch-dif9-staging.min.js"
  async
></script>
<script
  src="//test.url.com/?utm_source=newsletter&utm_medium=banner&utm_campaign=spring_sale"
  async
></script>
...
```

## Final Dockerfile
```Dockerfile=
FROM nginx:alpine

# copy nginx configuration
COPY nginx/default.conf /etc/nginx/conf.d/

# copy static content
COPY dist/ /html

# developers explicitly define and maintain VAR_NAMES_STORAGE in Dockerfile so that operations engineers could always use it as a reference and source of truth for the list of configuration parameters
ENV VAR_NAMES_STORAGE ADOBE_ANALYTICS_URL,GOOGLE_ANALYTICS_URL

# run sev and then run nginx
CMD ["/bin/sh", "-c", "sev /html && nginx -g 'daemon off;'"]
```

# Explaining `how sev works`
## Syntax
`sev` `/html/index.html`
* `/html/index.html` is `destination`
* `destination` can be `path to a file`
* or `path to a directory`
## Simple mode (example-1)
In this mode we simply replace `placeholders` `in files` with the corresponding `environment variables values`.

### Service environment variables
We have a `sevice environment variables` which are used to initialize `sev`.
#### VAR_NAMES_STORAGE
Stores the `list of environment variables names` which should be processed by `sev`.

##### Example
```sh=
export VAR_NAMES_STORAGE=ADOBE_ANALYTICS_URL,GOOGLE_ANALYTICS_URL
# you can separate environment variable names with comma
```

### Processing stage
* `sev` recursively processes each file provided in `destination`
* in each file `sev` replaces placeholders like **`_{ENVIRONMENT_VARIABLE}_`** with the corresponding value of `ENVIRONMENT_VARIABLE`

## Pipeline mode (example-2)
In this mode we provide a `JSON structure` which contains `sets of values of environment variables` segregated by `identifier`.

### Service environment variables
#### VAR_NAMES_STORAGE
Serves the same purpose as in `simple mode`
##### Example
```sh=
export environment-variable-4=some-value-4
export check=in_environment

export VAR_NAMES_STORAGE=environment-variable-4,check
```
#### VAR_VALUES_SETS_STORAGE
Stores the `JSON structure` which contains the `sets of values`.
##### Syntax
```json
{
  "set-of-values-identifier-1": {
    "variable-1": "the-value-v1-mk1"
  },
  "set-of-values-identifier-2": {
    "variable-1": "the-value-v1-mk2",
    "variable-2": "the-value-v2"
  },
  "set-of-values-identifier-3": {
    "variable-1": "the-value-v1-mk3",
    "variable-3": "the-value-v3"
  }
}
```
##### Example
```json=
{"akamai-prod":{"check":"in-structure","cntf_space":"9bnj36vfwq8e","cntf_token":"asd-AsdAsdAsdAsdAsdD-_AsdAsdAsdAsdAsd-As60","cntf_env":"master","chtr_log_level":0},"akamai":{"cntf_space":"asd234asd123asd","cntf_token":"asd-AsdAsdAsdAsdAsdD-_AsdAsdAsdAsdAsd-As60","cntf_env":"chatr-qa1","chtr_log_level":1},"dev":{"cntf_space":"asd456asd123asd","cntf_token":"asd-AsdAsdAsdAsdAsdD-_AsdAsdAsdAsdAsd-As60","cntf_env":"chatr-dev","chtr_log_level":1}}
```
#### VAR_VALUES_SETS_CHOSEN_ID
Contains the `identifier` for preferred `set of values`.
##### Example
```sh=
export VAR_VALUES_SETS_CHOSEN_ID=akamai-prod
```

#### Sample output
```sh
##### START SEV ANNOUNCEMENT #####

## Target environment variables
  environment-variable-4 = some-value-4
  check = in_environment


## Sets of values
► akamai-prod
  akamai
  dev


## Fetched values
  chtr_log_level = 0
  check = in-structure
  cntf_space = asd234asd123asd
  cntf_token = asd-AsdAsdAsdAsdAsdD-_AsdAsdAsdAsdAsd-As60
  cntf_env = master


## Mode
  Pulling sets of values for variables from JSON structure saved at {env.VAR_VALUES_SETS_STORAGE}

###### END SEV ANNOUNCEMENT ######

2020/04/21 15:01:00 INFO: sev succeeded: test-dir\sample_to_process
2020/04/21 15:01:00 INFO: sev succeeded: test-dir\sub\sample_to_process

Process finished with exit code 0

```

#### Important note
If you have `environment variable` intialized both in `JSON struct` and in the `environment` then the `value` taken from `environment` will prevail and get higher priority.