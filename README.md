# 2024 Dcard Backend Intern Homework

This Go project implements a simplified ad serving service with RESTful APIs to create and list advertisements based on specific conditions such as age, gender, country, and platform.

## Get Started

### Prerequisites

- Go 1.21.6 or higher
- PostgreSQL
- A `config.yaml` file configured with your database connection details.

1. Clone the repository:

    ```bash
    git clone git@github.com:z-huang/dcard-backend-homework-2024.git
    cd dcard-backend-homework-2024

    # Install the required Go modules
    go mod tidy 
    ```

2. Ensure your PostgreSQL database is running and accessible as per your `config.yaml` configuration.

3. Run the application
    ```bash
    go run main.go
    ```

The server will start listening on `0.0.0.0:8080`.

## Usage

### Create an Advertisement

Send a POST request to `/api/v1/ad` with the advertisement details in JSON format.

JSON structure:

- title
- startAt
- endAt
- conditions: an array with objects with following fields. All fields are optional.
    - ageStart
    - ageEnd
    - gender: an array. Accepted values: F and M, indicating female and male.
    - country: an array. The country codes inside should follow [ISO 3166-1](https://en.wikipedia.org/wiki/ISO_3166-1).
    - platform: an array. Accepted values: android, ios, web.

Example:
```bash
curl -X POST -H "Content-Type: application/json" \
    "http://localhost:8080/api/v1/ad" \
    --data '{
        "title": "Ad Sample",
        "startAt": "2024-01-01T00:00:00Z",
        "endAt": "2024-12-31T23:59:59Z",
        "conditions": [
            {
                "ageStart": 20,
                "ageEnd": 30,
                "gender": ["F"],
                "country": ["TW", "JP"],
                "platform": ["android", "ios"]
            }
        ]
    }'
```

### List advertisements

Send a GET request to `/api/v1/ad` with the conditions as query parameters to list active advertisements that match specific conditions.

Parameters:

- offset, limit: used for pagination
- age
- gender
- country
- platform

Example:

```bash
curl -X GET -H "Content-Type: application/json" \
  "http://localhost:8080/api/v1/ad?age=25&gender=F&country=TW&platform=ios"
```

## License
This project is licensed under the MIT License.