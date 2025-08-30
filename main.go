package main

import (
    "go-evermos/config"
    "go-evermos/internal/entities"
    "go-evermos/internal/handler"
    "go-evermos/pkg"

    "github.com/gofiber/fiber/v2"
)

func main() {
    // Init DB
    config.InitDB()

    // Auto migrate
    config.DB.AutoMigrate(
        &entities.User{},
		&entities.Category{},
		&entities.Store{},
		&entities.Product{},
		&entities.ProductLog{},
		&entities.ProductPicture{},
		&entities.Trx{},
		&entities.TrxDetail{},
		&entities.Address{},
    )

    app := fiber.New()

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("API Ecommerce Jalan")
    })
    
    // Public routes
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	// Protected routes
	user := app.Group("/user", pkg.JWTMiddleware())
	user.Get("/profile", handler.Profile)
    user.Put("/profile", handler.UpdateProfile) //update profil

    store := app.Group("/store", pkg.JWTMiddleware())
    store.Get("/", handler.GetMyStore)
    store.Put("/", handler.UpdateMyStore)

    address := app.Group("/address", pkg.JWTMiddleware())
    address.Post("/", handler.CreateAddress)
    address.Get("/", handler.GetAddresses)
    address.Put("/:id", handler.UpdateAddress)
    address.Delete("/:id", handler.DeleteAddress)

    category := app.Group("/categories", pkg.JWTMiddleware(), pkg.AdminOnly())
    category.Post("/", handler.CreateCategory)
    category.Get("/", handler.GetCategories)
    category.Put("/:id", handler.UpdateCategory)
    category.Delete("/:id", handler.DeleteCategory)

    product := app.Group("/product", pkg.JWTMiddleware())
    product.Post("/", handler.CreateProduct)
    app.Post("/product", handler.CreateProduct)
    app.Get("/products", handler.GetAllProducts)
    app.Get("/product/:id", handler.GetProductByID)
    app.Put("/product/:id", handler.UpdateProduct)
    app.Delete("/product/:id", handler.DeleteProduct)

    transaction := app.Group("/transactions", pkg.JWTMiddleware())
    transaction.Post("/", handler.CreateTransaction)
    transaction.Get("/", handler.GetUserTransactions)
    transaction.Get("/:id", handler.GetUserTransactionByID)



    app.Listen(":3000")
}
