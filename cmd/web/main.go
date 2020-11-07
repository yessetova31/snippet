package main
import (
	"aitu.com/snippet/pkg/models/postgres"

	"database/sql"
	"flag"
	_ "github.com/lib/pq"

	"fmt"
	"github.com/golangcollege/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	errorLog *log.Logger
	infoLog *log.Logger
	snippets *postgres.SnippetModel
	templateCache map[string]*template.Template
	session *sessions.Session

}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	db, err := sql.Open("postgresql", *dsn)
	if err != nil {
		errorLog.Fatal(err) }
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")

	if err != nil {
		errorLog.Fatal(err)
	}


	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		session: session,
		snippets: &sql.SnippetModel{DB: db},
		templateCache: templateCache,

	}
	srv := &http.Server{ Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}




func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.Header().Set("Connection", "close")
			app.serverError(w, fmt.Errorf("%s", err))
		} }()
	next.ServeHTTP(w, r) })
}
