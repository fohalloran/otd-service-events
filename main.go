package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type Event struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Distance              string    `json:"dist"`
	Location              string    `json:"location"`
	Ticket_price          string    `json:"ticket_price"`
	Ticket_sale_date_time time.Time `json:"ticket_sale_date_time"`
	Start_time            time.Time `json:"start_time"`
	Tickets_remaining     string    `json:"tickets_remaining"`
	Description     	  string    `json:"description"`
	Image_URL			  string	`json:"image_url"`
}

type NewEvent struct {
	Name                  string    `json:"name"`
	Location              string    `json:"location"`
	Ticket_price          string    `json:"ticket_price"`
	Start_time            string 	`json:"start_time"`
	Tickets_remaining     string    `json:"tickets_remaining"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}


func getEvents(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get events called")

	enableCors(&w)
	userLat := r.URL.Query().Get("latitude")
	userLong := r.URL.Query().Get("longitude")
	maxDistance := r.URL.Query().Get("maxDistance")
	orderBy := r.URL.Query().Get("orderBy")
	orderDirection := r.URL.Query().Get("orderDirection")
	maxCount := r.URL.Query().Get("maxCount")
	maxPrice := r.URL.Query().Get("maxPrice")
	whereAdded := false

	if userLat == "" || userLong == "" || maxDistance == "" {
		fmt.Println("Not all parameters given")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := os.Getenv("db_url")
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Close(context.Background())

	baseQuery := fmt.Sprintf("SELECT * FROM (SELECT events.id, events.name, locations.name, ticket_price, ticket_sale_date_time, start_time, tickets_remaining, image_url, dist FROM "+
		"(SELECT id, name, dist FROM (SELECT id, name, ST_Distance(ST_GeomFromText('POINT(%s %s)',4269),coords::GEOGRAPHY) AS dist "+
		"FROM locations) "+
		"WHERE dist < %s) as locations "+
		"INNER JOIN events ON locations.id = events.location_id)", userLong, userLat, maxDistance)

	if maxPrice != "" {
		if !whereAdded {
			baseQuery += " WHERE "
			whereAdded = true
		}
		baseQuery += fmt.Sprintf("ticket_price <= %s ", maxPrice)
	}

	if orderBy != "" {
		if orderDirection == "" {
			orderDirection = "ASC"
		}
		baseQuery += fmt.Sprintf("order by %s %s", orderBy, orderDirection)
	}

	if maxCount != "" {
		baseQuery += fmt.Sprintf(" limit %s", maxCount)
	}

	fmt.Println(baseQuery)

	rows, err := conn.Query(context.Background(), baseQuery)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var events []Event

	for rows.Next() {
		var newEvent Event
		err := rows.Scan(&newEvent.ID, &newEvent.Name, &newEvent.Location, &newEvent.Ticket_price, &newEvent.Ticket_sale_date_time, &newEvent.Start_time, &newEvent.Tickets_remaining,&newEvent.Image_URL, &newEvent.Distance)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		events = append(events, newEvent)
	}

	body, err := json.Marshal(events)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(body)

}

func getEventById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get event by Id called")
	enableCors(&w)
	eventId := r.PathValue("id")

	url := os.Getenv("db_url")
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Close(context.Background())
	query := fmt.Sprintf("SELECT events.id, events.name, events.description, locations.name, ticket_price, ticket_sale_date_time, start_time, image_url FROM (SELECT * FROM events where id='%s') as events INNER JOIN locations ON events.location_id = locations.id;", eventId)
	fmt.Println(query)
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		w.WriteHeader(500)
		return
	}

	var event Event

	for rows.Next() {
		err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.Ticket_price, &event.Ticket_sale_date_time, &event.Start_time,&event.Image_URL)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(event)
	}

	body, err := json.Marshal(event)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(body)

}

func addEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Add event called")

	url := os.Getenv("db_url")
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())


	decoder := json.NewDecoder(r.Body)
    var event NewEvent
    err = decoder.Decode(&event)
    if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid post body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
    }
	query := fmt.Sprintf("INSERT INTO events (name,ticket_price,tickets_remaining,location_id,start_time) VALUES ('%s','%s','%s','%s','%s','%s')",event.Name,event.Ticket_price,event.Tickets_remaining,event.Location,event.Start_time)
	fmt.Println(query)
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
    fmt.Println(event)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func main() {
	// my house: 51.465317091478745, -0.15783668426331612
	http.HandleFunc("GET /api/events/{id}", getEventById)
	http.HandleFunc("GET /api/events", getEvents)
	http.HandleFunc("POST /api/events", addEvent)
	http.ListenAndServe(":3000", nil)
}