# Property Listing System

A RESTful backend API service built with Go's Echo framework for managing property listings. Supports CRUD operations, advanced filtering, user authentication, property favorites, and recommendations. Uses MongoDB for data storage and Redis Cloud for caching.

## Project Requirements

Built to manage property listings from [dataset](https://cdn2.gro.care/db424fd9fb74_1748258398689.csv) with the following capabilities:

- **Data Import**: CSV dataset imported into MongoDB properties collection
- **CRUD Operations**: Create, read, update, delete properties with user-based access control
- **Advanced Filtering**: 10+ attributes including price, location, amenities, ratings
- **Caching**: Redis Cloud integration with 30-second TTL and cache invalidation
- **User Authentication**: JWT-based registration and login system
- **Favorites**: Users can manage favorite properties
- **Recommendations**: Property recommendation system between users
- **Deployment**: Dockerized and deployed on Render

## Features

- User registration/login with JWT authentication
- Property CRUD with creator-only update/delete permissions
- Admin superuser role for all operations
- Advanced filtering on 10+ attributes with pagination
- Favorite properties management per user
- Property recommendations via email search
- Redis Cloud caching for all read operations
- Dynamic cache keys using MD5 hashing
- Dockerized deployment on Render

## Tech Stack

- **Language**: Go (1.23)
- **Framework**: Echo (v4.13.4)
- **Database**: MongoDB Atlas
- **Caching**: Redis Cloud
- **Deployment**: Render (Docker)

**Dependencies**:
- `github.com/labstack/echo/v4`
- `go.mongodb.org/mongo-driver`
- `github.com/redis/go-redis/v9`
- `github.com/joho/godotenv`

## Project Structure

```
PropertyListingSys/
├── config/
│   └── database.go       # MongoDB connection setup
├── handlers/
│   ├── favorite.go       # Favorite CRUD handlers
│   ├── property.go       # Property CRUD and filter handlers
│   ├── recommendation.go # Recommendation handlers
│   └── user.go           # User auth and profile handlers
├── middleware/
│   └── jwt.go            # JWT authentication middleware
├── models/
│   ├── favorite.go       # Favorite model
│   ├── property.go       # Property model
│   ├── recommendation.go # Recommendation model
│   └── user.go           # User and auth request models
├── routes/
│   └── routes.go         # API route definitions
├── utils/
│   ├── redis.go          # Redis Cloud client and caching utilities
│   ├── jwt.go            # JWT generation and validation
│   └── password.go       # Password hashing and verification

├── Dockerfile            # Docker build configuration
├── go.mod                # Go module dependencies
├── main.go               # Application entry point
└── README.md             # Project documentation
```

## Prerequisites

- Go 1.23 or later
- MongoDB Atlas connection string
- Redis Cloud account with connection URL
- Docker (for local testing)

## Installation

1. **Clone the Repository**:
```bash
git clone https://github.com/Mayankrai449/PropertyListingSys.git
cd PropertyListingSys
```

2. **Install Dependencies**:
```bash
go mod tidy
go get github.com/labstack/echo/v4@latest
go get github.com/redis/go-redis/v9
```

3. **Import CSV Data**:
   - Download dataset: https://cdn2.gro.care/db424fd9fb74_1748258398689.csv
   - Import into MongoDB:
   ```bash
   mongoimport --uri "<MONGODB_URI>" --collection properties --type csv --headerline --file dataset.csv
   ```

4. **Set Up Environment Variables**:
   Create a `.env` file:
```env
MONGODB_URI=mongodb+srv://<user>:<pass>@<cluster>.mongodb.net/<db>?retryWrites=true&w=majority
MONGODB_DATABASE=property_db
MONGODB_COLLECTION_PROPERTIES=properties
MONGODB_COLLECTION_USER=user
MONGODB_COLLECTION_FAVORITES=favorites
MONGODB_COLLECTION_RECOMMENDATIONS=recommendations
REDIS_ADDR=redis://:<password>@<redis-cloud-host>:<port>
REDIS_PASSWORD=<redis-cloud-password>
PORT=8080
JWT_SECRET=your_jwt_secret
```

5. **Run Locally**:
```bash
go run main.go
```
Server runs at http://localhost:8080

## API Documentation

### List Properties (GET /properties)

Retrieves paginated property listings with advanced filtering. Responses cached for 30 seconds.

**Query Parameters**:
- `title` (string): Partial match on property title
- `type` (string): Property type (Villa, Apartment, etc.)
- `price_min`, `price_max` (float): Price range
- `state`, `city` (string): Location filters
- `area_min`, `area_max` (float): Area in square feet
- `bedrooms`, `bathrooms` (int): Room counts
- `amenities` (string): Pipe-separated amenities (pool|gym)
- `furnished` (string): Furnished status
- `available_from` (string): ISO date for availability
- `listed_by` (string): Listed by (Builder, Owner, etc.)
- `tags` (string): Pipe-separated tags (luxury|modern)
- `color_theme` (string): Color theme hex code
- `rating_min`, `rating_max` (float): Rating range (0–5)
- `is_verified` (bool): Verification status
- `listing_type` (string): Listing type (sale, rent)
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 10)

**Example Request**:
```
GET /properties?city=Mysore&price_min=20000000&bedrooms=4&page=1&limit=10
```

**Example Response**:
```json
[
  {
    "externalId": "PROP1001",
    "title": "Luxury Villa",
    "type": "Villa",
    "price": 25000000,
    "state": "Karnataka",
    "city": "Mysore",
    "areaSqFt": 3500,
    "bedrooms": 4,
    "bathrooms": 3,
    "createdAt": "2025-05-31T19:59:00+05:30"
  }
]
```

**Notes**:
- Cache key is MD5 hash of query parameters
- Filters are case-insensitive where applicable