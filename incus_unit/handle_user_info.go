package incus_unit

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "time"
    //for file io, net  connection

    "log"
    //for logging

    "go.mongodb.org/mongo-driver/bson"
    "golang.org/x/crypto/bcrypt"
    //for mongodb

    "github.com/gg582/linux_virt_unit"
    linux_virt_unit_crypto "github.com/gg582/linux_virt_unit/crypto"
    db "github.com/gg582/linux_virt_unit/mongo_connect"
    //custom packages
)

// CheckUserExists checks if a user exists in the database.
func CheckUserExists(username string, password string, ctx context.Context) bool {
    const maxWait = time.Second

    cursor, err := db.UserInfoCollection.Find(ctx, bson.D{})
    if err != nil {
        log.Printf("CheckUserExists: Failed to query users: %v", err)
        return false
    }
    defer cursor.Close(ctx)

    start := time.Now()
    for cursor.Next(ctx) {
        var user linux_virt_unit.UserInfo
        if err := cursor.Decode(&user); err != nil {
            continue
        }
        if user.Username == username {
            err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
            if err != nil {
                log.Printf("CheckUserExists: Password mismatch for user '%s'", username)
                return false
            }
            return true
        }
        if time.Since(start) > maxWait {
            log.Printf("CheckUserExists: Timed out while scanning for user '%s'", username)
            return false
        }
    }

    log.Printf("CheckUserExists: User '%s' not found", username)
    return false
}

// Register godoc
// @Summary Register a new user
// @Description Registers a new user
// @Accept json
// @Produce json
// @Param request body linux_virt_unit.UserInfo true "User registration request"
// @Success 200 body string true "User Registration Done."
// @Failure 400
// @Router /register [post]
func Register(wr http.ResponseWriter, req *http.Request) {
    if req.Method != http.MethodPost {
        http.Error(wr, "This endpoint allows only POST methods. aborting", http.StatusMethodNotAllowed)
        return
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var u linux_virt_unit.UserInfo
    body, err := io.ReadAll(req.Body)
    if err != nil {
        log.Printf("Register: Failed to read request body: %v", err)
        http.Error(wr, "Failed to read request body", http.StatusBadRequest)
        return
    }

    // Unmarshal the JSON into UserInfo struct
    if err := json.Unmarshal(body, &u); err != nil {
        log.Printf("Register: Invalid JSON: %v", err)
        http.Error(wr, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Decrypt the username from client
    username, err := linux_virt_unit_crypto.DecryptString(u.Username, u.Key, u.UsernameIV)
    if err != nil {
        log.Printf("Register: Failed to decrypt username: %v", err)
        http.Error(wr, "Username decryption failed", http.StatusBadRequest)
        return
    }

    // Check if the username already exists
    if CheckUserExists(username, u.Password, ctx) {
        log.Printf("Register: Username '%s' already exists", username)
        http.Error(wr, "User already exists", http.StatusConflict)
        return
    }

    // Prepare the user struct for DB insertion
    u.Username = username

    // Insert the new user into MongoDB
    if _, err := db.UserInfoCollection.InsertOne(ctx, u); err != nil {
        log.Printf("Register: Failed to insert user into DB: %v", err)
        http.Error(wr, "Failed to register user", http.StatusInternalServerError)
        return
    }

    log.Printf("Register: User '%s' registered successfully", username)
    wr.Write([]byte("User Registration Done"))
}
