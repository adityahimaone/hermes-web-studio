package auth

import "testing"

func TestPasswordSetupLoginAndSignedCookie(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if s.Enabled() {
		t.Fatal("new auth should be disabled")
	}
	if err := s.Setup("correct horse battery"); err != nil {
		t.Fatal(err)
	}
	token, err := s.Login("127.0.0.1", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !s.Verify(token) || s.Verify(token+"x") {
		t.Fatal("cookie signature validation failed")
	}
	if _, err := s.Login("127.0.0.1", "wrong"); err == nil {
		t.Fatal("wrong password accepted")
	}
}

func TestLoginRateLimit(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Setup("correct horse battery"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		_, _ = s.Login("127.0.0.2", "wrong")
	}
	if _, err := s.Login("127.0.0.2", "wrong"); err != ErrRateLimited {
		t.Fatalf("err=%v", err)
	}
}

func TestLoginRotatesCookieEvenWithinOneSecond(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Setup("correct horse battery"); err != nil {
		t.Fatal(err)
	}
	first, err := s.Login("127.0.0.1", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Login("127.0.0.1", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !s.Verify(first) || !s.Verify(second) {
		t.Fatalf("expected distinct valid rotated cookies, first=%q second=%q", first, second)
	}
}
