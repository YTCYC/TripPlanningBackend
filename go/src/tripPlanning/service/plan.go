/*
	 Manage user's plans, including:
		generate a plan from selected places and save it to database
		delete a plan
		read user's plans
*/
package service

import (
	"log"
	"strings"
	"tripPlanning/backend"
	"tripPlanning/model"

	"github.com/pborman/uuid"
)

func GeneratePlanAndSaveToDB(userID string, placesOfAllDays [][]model.Place,
	startDay string, endDay string, transportation string, tripName string) error {
	// params:
	// placesOfAllDays: Each sub-array represent the planned places to visit each day

	// 1. create a new TirpPlan for this user
	tripID := uuid.New()
	tripTableEntry := map[string]interface{}{
		"tripID":         tripID,
		"userID":         userID,
		"tripName":       tripName,
		"startDay":       startDay,
		"endDay":         endDay,
		"transportation": transportation,
	}
	err := backend.InsertIntoDB(backend.TableName_Trips, tripTableEntry)
	if err != nil {
		log.Fatal("Error during store new trip plan: ", err)
		return err
	}

	// 2. plan route for each day,
	var plannedRoutes [][]model.Place
	for _, placesEachDay := range placesOfAllDays {
		sortedPlaces, err := generateDayPlan(placesEachDay, transportation, "")
		if err != nil {
			log.Fatal("Error during sorting places for a day: ", err)
			return err
		}
		plannedRoutes = append(plannedRoutes, sortedPlaces)
	}

	// 3. save planned routes to db
	for dayOrder, planedRoute := range plannedRoutes {
		// 3.1 save each dayPlan to DB
		currentDayPlanId := uuid.New()
		tripTableEntry := map[string]interface{}{
			"dayPlanID": currentDayPlanId,
			"tripID":    tripID,
			"dayOrder":  dayOrder + 1,
		}
		err = backend.InsertIntoDB(backend.TableName_DayPlans, tripTableEntry)
		if err != nil {
			log.Fatal("Error during store new day-plan: ", err)
			return err
		}

		// 3.2 save each places of the day
		for visitOrder, place := range planedRoute {
			// 3.2.1 save the place detail if necessary
			placeID := place.Id
			// TODO: check if the placeID is already in DB, if not then save it to DB
			placeIsInDB := false
			if !placeIsInDB {
				err = savePlaceToDB(place)
				if err != nil {
					log.Fatal("Error during store new trip place: ", err)
					return err
				}
			}
			// 3.2.1 save the day-place relation
			dayPlaceRelationEntry := map[string]interface{}{
				"placeID":    placeID,
				"dayPlanID":  currentDayPlanId,
				"visitOrder": visitOrder + 1,
			}
			err = backend.InsertIntoDB(backend.TableName_DayPlaceRelations, dayPlaceRelationEntry)
			if err != nil {
				log.Fatal("Error during store new day-place relation: ", err)
				return err
			}
		}
	}
	return nil
}

func generateDayPlan(places []model.Place, transportation string, date string) ([]model.Place, error) {
	// THIS IS NOT DONE YET
	return places, nil
}

func savePlaceToDB(place model.Place) error {
	// save place
	var photoURLs []string
	for _, p := range place.Photos {
		photoURLs = append(photoURLs, p.Id)
	}
	placeEntry := map[string]interface{}{
		"placeID":   place.Id,
		"name":      place.DisplayName.Text,
		"address":   place.Address,
		"placeType": place.PlaceType.Text,
		"photoURLs": strings.Join(photoURLs, "$$"),
	}

	err := backend.InsertIntoDB(backend.TableName_PlaceDetails, placeEntry)
	if err != nil {
		log.Println("Error during store new place: ", err)
		return err
	}
	// save reviews of this place
	for _, review := range place.Reviews {
		reviewEntry := map[string]interface{}{
			"reviewID":    uuid.New(),
			"reviewText":  review.Text.Text,
			"rating":      review.Rating,
			"publishTime": review.PublishTime,
			"placeID":     place.Id,
		}
		err = backend.InsertIntoDB(backend.TableName_Reviews, reviewEntry)
		if err != nil {
			log.Fatal("Error during store new review: ", err)
			return err
		}
	}
	return nil
}
