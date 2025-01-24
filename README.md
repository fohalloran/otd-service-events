# Event Service for On The Door  

This repository contains the backend service for the "On The Door" application. The Event Service provides APIs for managing events, including retrieving event details, searching for events near a location, and adding new events to the system.  

---

## Features  

- **Get Events**: Retrieve a list of events filtered by location, distance, price, and other parameters.  
- **Get Event by ID**: Fetch detailed information about a specific event.  
- **Add Event**: Add new events to the database.  

---

## API Endpoints  

### 1. **Get Events**  
- **Endpoint**: `GET /api/events`  
- **Query Parameters**:  
  - `latitude`: User's latitude (required).  
  - `longitude`: User's longitude (required).  
  - `maxDistance`: Maximum distance from the user in meters (required).  
  - `orderBy`: Column to order by (e.g., `ticket_price`, `start_time`) (optional).  
  - `orderDirection`: Sorting direction (`ASC` or `DESC`) (optional).  
  - `maxCount`: Maximum number of results to return (optional).  
  - `maxPrice`: Maximum ticket price (optional).  
- **Response**: A JSON array of events with details like name, location, ticket price, and more.  

### 2. **Get Event by ID**  
- **Endpoint**: `GET /api/events/{id}`  
- **Path Parameter**:  
  - `id`: Event ID (required).  
- **Response**: A JSON object containing detailed information about the event.  

### 3. **Add Event**  
- **Endpoint**: `POST /api/events`  
- **Request Body**:  
  ```json
  {
    "name": "Event Name",
    "location": "Location Name",
    "ticket_price": "Price in Currency",
    "start_time": "Start Time in ISO Format",
    "tickets_remaining": "Number of Tickets Remaining"
  }
