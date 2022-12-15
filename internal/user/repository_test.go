//go:build integration
// +build integration

package user

import (
	"faceit/internal/common"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type (
	repositoryTestSuite struct {
		repo gormRepository
		suite.Suite
	}
)

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(repositoryTestSuite))
}

func (s *repositoryTestSuite) SetupSuite() {
	r, err := NewRepository()
	s.Require().NoError(err)
	s.repo = *r
}

func (s repositoryTestSuite) TestFindByID() {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	actualUser, err := s.repo.findByID(nil, id)

	s.NoError(err)
	s.Equal("John", actualUser.FirstName)
	s.Equal("Doe", actualUser.LastName)
	s.Equal("johndoe", actualUser.Nickname)
	s.Equal("johndoe@email.com", actualUser.Email)
	s.Equal("US", actualUser.Country)
}

func (s repositoryTestSuite) TestFindByID_ReturnsNotFound() {
	_, err := s.repo.findByID(nil, uuid.Nil)
	s.ErrorIs(err, ErrUserNotFound)
}

func (s repositoryTestSuite) TestListPagination() {
	s.reinitDB()

	res, err := s.repo.list(nil, common.Pagination{Page: 0, PageSize: 2}, nil)
	s.NoError(err)
	s.Len(res, 2)
	s.Equal("dome@email.com", res[0].Email)
	s.Equal("janedoe@email.com", res[1].Email)

	res, err = s.repo.list(nil, common.Pagination{Page: 2, PageSize: 1}, nil)
	s.NoError(err)
	s.Len(res, 1)
	s.Equal("johndoe@email.com", res[0].Email)

	res, err = s.repo.list(nil, common.Pagination{}, nil)
	s.NoError(err)
	s.Len(res, 3)
}

func (s repositoryTestSuite) TestListFilter() {
	s.reinitDB()
	p := common.Pagination{Page: 0, PageSize: 0}

	for _, test := range []struct {
		name           string
		filter         User
		expectedEmails []string
	}{
		{
			name:           "first name",
			filter:         User{FirstName: "jo"},
			expectedEmails: []string{"johndoe@email.com"},
		},
		{
			name:           "last name",
			filter:         User{LastName: "doe"},
			expectedEmails: []string{"janedoe@email.com", "johndoe@email.com"},
		},
		{
			name:           "nickname",
			filter:         User{Nickname: "j"},
			expectedEmails: []string{"janedoe@email.com", "johndoe@email.com"},
		},
		{
			name:           "email",
			filter:         User{Email: "do"},
			expectedEmails: []string{"dome@email.com"},
		},
		{
			name:           "country",
			filter:         User{Country: "uk"},
			expectedEmails: []string{"dome@email.com", "janedoe@email.com"},
		},
		{
			name:           "last name and country",
			filter:         User{LastName: "do", Country: "uk"},
			expectedEmails: []string{"dome@email.com", "janedoe@email.com"},
		},
		{
			name:           "first name and different country",
			filter:         User{FirstName: "zoltan", Country: "us"},
			expectedEmails: []string{},
		},
	} {
		s.Run(test.name, func() {
			res, err := s.repo.list(nil, p, &test.filter)
			s.NoError(err)
			s.Len(res, len(test.expectedEmails))

			actualEmails := []string{}
			for _, r := range res {
				actualEmails = append(actualEmails, r.Email)
			}
			s.Equal(test.expectedEmails, actualEmails)
		})
	}
}

func (s repositoryTestSuite) TestDeleteByID() {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	s.NoError(s.repo.db.Create(&User{
		ID:        id,
		FirstName: "fn",
		LastName:  "ln",
		Nickname:  "nn",
		Email:     "email",
		Country:   "US",
	}).Error)

	s.NoError(s.repo.deleteByID(nil, id))

	s.ErrorIs(s.repo.db.Take(&User{}, id).Error, gorm.ErrRecordNotFound)
}

func (s repositoryTestSuite) TestCreate() {
	user := User{
		FirstName: "create-fn",
		LastName:  "create-ln",
		Nickname:  "create-nn",
		Email:     "email",
		Country:   "US",
	}

	newUser, err := s.repo.create(nil, user, "testpwd")
	s.NoError(err)
	s.NotNil(newUser.ID)
	s.True(newUser.ID != uuid.Nil)
	s.True(time.Now().Sub(newUser.CreatedAt) < time.Second)
	s.True(time.Now().Sub(*newUser.UpdatedAt) < time.Second)

	var savedPwds []string
	s.NoError(s.repo.db.Model(User{}).Where("id = ?", newUser.ID).Pluck("password", &savedPwds).Error)
	s.Equal("testpwd", savedPwds[0])

	s.repo.db.Delete(&User{}, newUser.ID)
}

func (s repositoryTestSuite) TestUpdatePassword() {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	pwd := uuid.New().String()[0:4]

	s.NoError(s.repo.updatePassword(nil, id, pwd))

	var savedPwds []string
	s.Require().NoError(s.repo.db.Model(User{}).Where("id = ?", id).Pluck("password", &savedPwds).Error)
	s.Equal(pwd, savedPwds[0])
}

func (s repositoryTestSuite) TestUpdate() {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	originalUser := User{
		FirstName: "Zoltan",
		LastName:  "Domahidi",
		Nickname:  "dome",
		Email:     "dome@email.com",
		Country:   "UK",
	}
	getChangeAndExpected := func(change func(*User)) (User, User) {
		changeUser := User{}
		change(&changeUser)

		originalUserCopy := originalUser
		change(&originalUserCopy)
		return changeUser, originalUserCopy
	}

	for _, change := range []func(*User){
		func(u *User) { u.FirstName = "newFirstName" },
		func(u *User) { u.LastName = "newLastName" },
		func(u *User) { u.Nickname = "newNickname" },
		func(u *User) { u.Email = "newEmail" },
		func(u *User) { u.Country = "HU" },
	} {
		s.reinitDB()
		change, expected := getChangeAndExpected(change)
		updatedUser, err := s.repo.update(nil, id, change)
		s.NoError(err)
		s.Equal(expected.FirstName, updatedUser.FirstName)
		s.Equal(expected.LastName, updatedUser.LastName)
		s.Equal(expected.Nickname, updatedUser.Nickname)
		s.Equal(expected.Email, updatedUser.Email)
		s.Equal(expected.Country, updatedUser.Country)

	}
}

func (s repositoryTestSuite) reinitDB() {
	s.NoError(s.repo.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error)

	sql, err := os.ReadFile("../../scripts/initdb.sql")
	s.NoError(err)

	s.NoError(s.repo.db.Exec(string(sql)).Error)
}
