package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikitanaumenko/calendar/internal/api/gen"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

func TestRouterServesAPIUnderAPIPrefix(t *testing.T) {
	router := NewRouter(zap.NewNop(), stubAPI{}, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/event-types", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected prefixed API route to reach API handler, got status %d", response.Code)
	}
}

func TestRouterServesSPAFallbackFromStaticDir(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(staticDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html><title>Calendar</title>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "assets", "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(zap.NewNop(), stubAPI{}, staticDir)

	for _, path := range []string{"/", "/booking"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected %s to serve SPA index, got status %d", path, response.Code)
		}
		if !strings.Contains(response.Body.String(), "<title>Calendar</title>") {
			t.Fatalf("expected %s to serve index.html, got %q", path, response.Body.String())
		}
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected asset to be served, got status %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "console.log('ok')") {
		t.Fatalf("expected asset body, got %q", response.Body.String())
	}
}

func TestRouterDoesNotServeSPAForUnknownAPIPath(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<!doctype html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(zap.NewNop(), stubAPI{}, staticDir)

	for _, path := range []string{"/api", "/api/unknown"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected %s to return API 404, got status %d", path, response.Code)
		}
	}
}

type stubAPI struct{}

func (stubAPI) AdminBookingsListUpcomingBookings(context.Context, gen.AdminBookingsListUpcomingBookingsRequestObject) (gen.AdminBookingsListUpcomingBookingsResponseObject, error) {
	return gen.AdminBookingsListUpcomingBookings200JSONResponse{}, nil
}

func (stubAPI) AdminEventTypesCreateEventType(context.Context, gen.AdminEventTypesCreateEventTypeRequestObject) (gen.AdminEventTypesCreateEventTypeResponseObject, error) {
	return gen.AdminEventTypesCreateEventType201JSONResponse{}, nil
}

func (stubAPI) BookingsCreateBooking(context.Context, gen.BookingsCreateBookingRequestObject) (gen.BookingsCreateBookingResponseObject, error) {
	return gen.BookingsCreateBooking201JSONResponse{}, nil
}

func (stubAPI) EventTypesListEventTypes(context.Context, gen.EventTypesListEventTypesRequestObject) (gen.EventTypesListEventTypesResponseObject, error) {
	return gen.EventTypesListEventTypes200JSONResponse{}, nil
}

func (stubAPI) SlotsListAvailableSlots(context.Context, gen.SlotsListAvailableSlotsRequestObject) (gen.SlotsListAvailableSlotsResponseObject, error) {
	return gen.SlotsListAvailableSlots200JSONResponse{
		EventTypeId: openapi_types.UUID{},
		Slots:       []gen.AvailableSlot{},
	}, nil
}
