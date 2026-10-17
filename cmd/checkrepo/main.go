package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"price-bot/internal/config"
	"price-bot/internal/storage"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := storage.Connect(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	users := storage.NewUserRepo(pool)
	items := storage.NewItemRepo(pool)

	fmt.Println("=== 1. EnsureUser (new) ===")
	u, err := users.EnsureUser(ctx, 11111, "test_user")
	fmt.Printf("u=%+v\nerr=%v\n\n", u, err)
	if err != nil {
		log.Fatalf("cannot continue: %v", err)
	}

	fmt.Println("=== 2. EnsureUser (again, rename) ===")
	u2, err := users.EnsureUser(ctx, 11111, "test_user_renamed")
	fmt.Printf("u2=%+v\nerr=%v\nsameID=%v\n\n", u2, err, u2 != nil && u.ID == u2.ID)

	fmt.Println("=== 3. GetByTgID (existing) ===")
	u3, err := users.GetByTgID(ctx, 11111)
	fmt.Printf("u3=%+v\nerr=%v\n\n", u3, err)

	fmt.Println("=== 4. GetByTgID (missing) ===")
	_, err = users.GetByTgID(ctx, 99999)
	fmt.Printf("err=%v\nisNotFound=%v\n\n", err, errors.Is(err, storage.ErrUserNotFound))

	fmt.Println("=== 5. EnsureUser (second user) ===")
	u4, err := users.EnsureUser(ctx, 22222, "user2")
	fmt.Printf("u4=%+v\nerr=%v\n\n", u4, err)
	if err != nil {
		log.Fatalf("cannot continue: %v", err)
	}

	fmt.Println("=== 6. Add item1 for user1 ===")
	it1, err := items.Add(ctx, u.ID, "https://example.com/p1", 99.99, "Product 1")
	fmt.Printf("it1=%+v\nerr=%v\n\n", it1, err)
	if err != nil {
		log.Fatalf("cannot continue: %v", err)
	}

	fmt.Println("=== 7. Add item2 for user1 ===")
	it2, err := items.Add(ctx, u.ID, "https://example.com/p2", 199.99, "Product 2")
	fmt.Printf("it2=%+v\nerr=%v\n\n", it2, err)

	fmt.Println("=== 8. Add duplicate (same user, same url) ===")
	_, err = items.Add(ctx, u.ID, "https://example.com/p1", 50.00, "Duplicate")
	fmt.Printf("err=%v\nisDuplicate=%v\n\n", err, errors.Is(err, storage.ErrDuplicate))

	fmt.Println("=== 9. Add item for user2 ===")
	_, err = items.Add(ctx, u4.ID, "https://example.com/p3", 300.00, "Product 3")
	fmt.Printf("err=%v\n\n", err)

	fmt.Println("=== 10. ListByUser user1 ===")
	list1, err := items.ListByUser(ctx, u.ID)
	fmt.Printf("count=%d\nerr=%v\n", len(list1), err)
	for _, it := range list1 {
		fmt.Printf("  id=%d url=%s target=%.2f title=%s\n", it.ID, it.URL, it.TargetPrice, it.Title)
	}
	fmt.Println()

	fmt.Println("=== 11. ListByUser user2 ===")
	list2, err := items.ListByUser(ctx, u4.ID)
	fmt.Printf("count=%d\nerr=%v\n", len(list2), err)
	for _, it := range list2 {
		fmt.Printf("  id=%d url=%s target=%.2f title=%s\n", it.ID, it.URL, it.TargetPrice, it.Title)
	}
	fmt.Println()

	fmt.Println("=== 12. CountByUser user1 (expected 2) ===")
	c1, err := items.CountByUser(ctx, u.ID)
	fmt.Printf("count=%d\nerr=%v\n\n", c1, err)

	fmt.Println("=== 13. CountByUser user2 (expected 1) ===")
	c2, err := items.CountByUser(ctx, u4.ID)
	fmt.Printf("count=%d\nerr=%v\n\n", c2, err)

	fmt.Println("=== 14. Remove item1 (existing) ===")
	err = items.Remove(ctx, u.ID, it1.ID)
	fmt.Printf("err=%v\n\n", err)

	fmt.Println("=== 15. Remove item1 again (should be not found) ===")
	err = items.Remove(ctx, u.ID, it1.ID)
	fmt.Printf("err=%v\nisNotFound=%v\n\n", err, errors.Is(err, storage.ErrItemNotFound))

	fmt.Println("=== 16. Remove foreign item (user2 removing user1's item2) ===")
	err = items.Remove(ctx, u4.ID, it2.ID)
	fmt.Printf("err=%v\nisNotFound=%v\n(should be NotFound, user2 doesn't own it)\n\n", err, errors.Is(err, storage.ErrItemNotFound))

	fmt.Println("=== 17. ListByUser user1 after remove (expected 1) ===")
	list1, err = items.ListByUser(ctx, u.ID)
	fmt.Printf("count=%d\nerr=%v\n\n", len(list1), err)

	fmt.Println("=== 18. Cleanup ===")
	// удаляем всех тестовых пользователей — каскадом уйдут items
	_, err = pool.Exec(ctx, "DELETE FROM users WHERE tg_id IN (11111, 22222)")
	fmt.Printf("cleanup err=%v\n\n", err)

	fmt.Println("=== ALL CHECKS DONE ===")
}
