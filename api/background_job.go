package api

import "fmt"

func (app *KeyKeeper) backgroundJob(fn func ()) {
	app.wg.Add(1)

	// Launch the background goroutine.
	go func ()  {
		defer app.wg.Done()

		defer func ()  {
			if err := recover(); err != nil {
				app.logger.Error().Err(fmt.Errorf("%s", err))
			}
		}()

		fn()
	}()
}