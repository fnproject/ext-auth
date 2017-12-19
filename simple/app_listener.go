// Deals with connecting apps and users.

package simple

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/fnproject/fn/api/common"
	"github.com/fnproject/fn/api/models"
)

// listener to be used as an Applistener on a functions server.
type listener struct {
	// store *store
	// cache *cache.Cache
	simple *SimpleAuth
}

// BeforeAppCreate is called before app create to add entry for app for the current user.
func (l *listener) BeforeAppCreate(ctx context.Context, app *models.App) error { return nil }

// AfterAppCreate we insert the user->app mapping
func (l *listener) AfterAppCreate(ctx context.Context, app *models.App) error {
	userID := ctx.Value(userIDKey).(string)
	_, err := l.simple.ds.GetDatabase().Exec(l.simple.insertUserAppQuery, userID, app.Name)
	if err != nil {
		// TODO: Now we have an app without an owner. Bad.
		common.Logger(ctx).WithError(err).Errorln("Failed to create user app association.")
		return authErr{http.StatusInternalServerError, "user app creation failed", nil}
	}
	return nil
}

// BeforeAppUpdate ...
func (l *listener) BeforeAppUpdate(ctx context.Context, app *models.App) error {
	userID := ctx.Value(userIDKey).(string)
	return l.simple.canAccessApp(ctx, userID, app.Name)
}

// AfterAppUpdate ...
func (l *listener) AfterAppUpdate(ctx context.Context, app *models.App) error { return nil }

// BeforeAppDelete ...
func (l *listener) BeforeAppDelete(ctx context.Context, app *models.App) error {
	userID := ctx.Value(userIDKey).(string)
	return l.simple.canAccessApp(ctx, userID, app.Name)
}

// AfterAppDelete ...
func (l *listener) AfterAppDelete(ctx context.Context, app *models.App) error {
	userID := ctx.Value(userIDKey).(string)
	// log := common.Logger(ctx)
	// remove user mapping
	_, err := l.simple.ds.GetDatabase().Exec(l.simple.deleteUserAppQuery, userID, app.Name)
	if err != nil {
		return authErr{http.StatusInternalServerError, "user app deletion failed", nil}
	}
	return nil
}

// BeforeAppGet called right before getting an app
func (l *listener) BeforeAppGet(ctx context.Context, appName string) error {
	userID := ctx.Value(userIDKey).(string)
	return l.simple.canAccessApp(ctx, userID, appName)
}

// AfterAppGet called after getting app from database
func (l *listener) AfterAppGet(ctx context.Context, app *models.App) error { return nil }

// BeforeAppsList called right before getting a list of all user's apps. Modify the filter to adjust what gets returned.
func (l *listener) BeforeAppsList(ctx context.Context, filter *models.AppFilter) error {
	// QUERY FOR USER APPS MAPPING, GET IDS, SET IN APPFILTER
	v := ctx.Value(userIDKey)
	userID := v.(string)
	var uapps []*UserApps
	err := l.simple.ds.GetDatabase().Select(&uapps, l.simple.getUserAppsByUserID, userID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	appNames := []string{}
	// appsMap := make(map[string]*UserApps)
	for _, uapp := range uapps {
		// appsMap[uapp.AppName] = uapp
		appNames = append(appNames, uapp.AppName)
	}
	// return appsMap, err
	filter.NameIn = appNames
	return nil
}

// AfterAppsList called after deleting getting a list of user's apps. apps is the result after applying AppFilter.
func (l *listener) AfterAppsList(ctx context.Context, apps []*models.App) error { return nil }
