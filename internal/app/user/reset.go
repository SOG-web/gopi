package user

// ResetPassword sets a new password for the given user ID without requiring the old password.
// Intended for use by the password reset flow after token verification.
func (s *UserService) ResetPassword(userID, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	hashed, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(user.ID, hashed)
}
