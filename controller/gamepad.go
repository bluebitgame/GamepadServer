package controller

func (r *Xbox360ControllerReport) SetButtonFromGamePad(button int, value byte) {
	switch button {
	case 0:
		r.MaybeSetButton(Xbox360ControllerButtonA, value != 0)
	case 1:
		r.MaybeSetButton(Xbox360ControllerButtonB, value != 0)
	case 2:
		r.MaybeSetButton(Xbox360ControllerButtonX, value != 0)
	case 3:
		r.MaybeSetButton(Xbox360ControllerButtonY, value != 0)
	case 4:
		r.MaybeSetButton(Xbox360ControllerButtonLeftShoulder, value != 0)
	case 5:
		r.MaybeSetButton(Xbox360ControllerButtonRightShoulder, value != 0)
	case 6:
		r.SetLeftTrigger(value)
	case 7:
		r.SetRightTrigger(value)
	case 8:
		r.MaybeSetButton(Xbox360ControllerButtonBack, value != 0)
	case 9:
		r.MaybeSetButton(Xbox360ControllerButtonStart, value != 0)
	case 10:
		r.MaybeSetButton(Xbox360ControllerButtonLeftThumb, value != 0)
	case 11:
		r.MaybeSetButton(Xbox360ControllerButtonRightThumb, value != 0)
	case 12:
		r.MaybeSetButton(Xbox360ControllerButtonUp, value != 0)
	case 13:
		r.MaybeSetButton(Xbox360ControllerButtonDown, value != 0)
	case 14:
		r.MaybeSetButton(Xbox360ControllerButtonLeft, value != 0)
	case 15:
		r.MaybeSetButton(Xbox360ControllerButtonRight, value != 0)
	case 16:
		r.MaybeSetButton(Xbox360ControllerButtonGuide, value != 0)
	}
}
