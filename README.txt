================================================================================
                                  CHATROOM
                        Forum — Documentation Technique
                  Projet ChatRoom · Go + SQLite + HTML/CSS/JS
================================================================================


================================================================================
 1. PRÉSENTATION DU PROJET
================================================================================
ChatRoom est un forum web complet développé en Go avec une base de données SQLite. 
Le projet reprend les caractéristiques d'un forum classique : 
publication de posts, réactions (like/dislike), système d'authentification, catégories, filtrage, recherche, profils utilisateurs et système de citations.


--- Objectifs pédagogiques ---
• Maîtriser un serveur web Go (net/http) sans framework
• Gérer une base de données SQLite avec le driver go-sqlite3 (CGO)
• Implémenter une authentification sécurisée avec bcrypt et sessions UUID
• Construire des templates HTML avec html/template (héritage de layouts)
• Produire un design sobre et professionnel sans framework CSS


--- Stack technique ---
| Couche           | Technologie                                                  |
|------------------+--------------------------------------------------------------|
| Backend          | Go 1.21+ · net/http · html/template                          |
| Base de données  | SQLite3 via github.com/mattn/go-sqlite3                      |
| Authentification | bcrypt (golang.org/x/crypto) + UUID (github.com/google/uuid) |
| Frontend         | HTML5 · CSS3 (variables, grid, flexbox) · JS vanilla         |
| Typographie      | Montserrat (Google Fonts)                                    |
--------------------------------------------------------------------------------

================================================================================
 2. ARCHITECTURE DU PROJET
================================================================================

--- Arborescence des fichiers ---
    forum/
    ├── main.go                    ← point d'entrée, routing
    ├── go.mod / go.sum
    ├── forum.db                   ← base SQLite (générée au premier lancement)
    ├── database/
    │   ├── db.go                  ← connexion + toutes les fonctions SQL
    │   └── schema.sql             ← définition des tables + données initiales
    ├── handlers/
    │   ├── auth.go                ← register, login, logout
    │   ├── topics.go              ← home, topic view, create topic, search, searchUsers
    │   ├── posts.go               ← create post, react
    │   ├── profile.go             ← profil utilisateur, mise à jour profil + avatar
    │   └── helpers.go             ← renderTemplate (map de templates), parseMentions
    ├── middleware/
    │   └── auth.go                ← Auth(), RequireAuth(), GetUserFromRequest()
    ├── models/
    │   └── models.go              ← structs Go (User, Topic, Post, FavoriteTopic, etc.)
    ├── static/
    │   ├── css/style.css
    │   ├── js/main.js
    │   ├── uploads/               ← avatars uploadés (créé automatiquement)
    │   └── ChatRoomLogo.png
    └── templates/
        ├── base.html              ← layout principal (header, nav, modal profil)
        ├── index.html             ← liste des topics + sidebar + recherche
        ├── topic.html             ← vue d'un topic avec posts, réactions et citations
        ├── login.html
        ├── register.html
        ├── create_topic.html
        └── profile.html          ← page profil utilisateur (avatar, bio, sujets préférés)


--- Flux d'une requête ---
Navigateur → main.go (routing) → middleware.Auth/RequireAuth → handler → database → renderTemplate → html/template → réponse HTML
--------------------------------------------------------------------------------

================================================================================
 3. BASE DE DONNÉES
================================================================================

--- Initialisation ---
Au démarrage, db.go ouvre (ou crée) forum.db, active les clés étrangères via PRAGMA foreign_keys = ON (désactivées par défaut dans SQLite), puis exécute schema.sql. Grâce aux clauses IF NOT EXISTS et INSERT OR IGNORE, cette opération est idempotente : les données existantes ne sont jamais écrasées.


--- Tables ---
| Table      | Rôle et colonnes clés                                                              |
|------------+------------------------------------------------------------------------------------|
| users      | id, username (UNIQUE), email (UNIQUE), password_hash, avatar_url, bio, created_at  |
| sessions   | uuid (PK), user_id (FK), expires_at — une session = un cookie actif                |
| categories | id, name (UNIQUE), slug (UNIQUE) — le slug sert aux URLs propres                   |
| topics     | id, user_id (FK), category_id (FK), title, description, created_at                 |
| posts      | id, topic_id (FK), user_id (FK), content, image_url, parent_id, created_at         |
| reactions  | user_id + post_id (UNIQUE composite) + type CHECK('like','dislike')                |

Notes sur les colonnes ajoutées :
• users.avatar_url (TEXT NOT NULL DEFAULT '') : URL de la photo de profil (chemin /static/uploads/ après upload)
• users.bio (TEXT NOT NULL DEFAULT '') : texte libre de présentation
• topics.description (TEXT NOT NULL DEFAULT '') : description courte optionnelle à la création du sujet
• posts.image_url (TEXT DEFAULT '') : URL d'une image attachée au post
• posts.parent_id (INTEGER REFERENCES posts(id)) : référence au post cité (système de citations)


--- Contraintes importantes ---
• UNIQUE(user_id, post_id) dans reactions : empêche physiquement un double vote, même en cas de bug applicatif.
• CHECK(type IN ('like','dislike')) : la valeur est validée en BDD, pas seulement côté Go.
• ON DELETE CASCADE sur posts et reactions : supprimer un topic supprime automatiquement ses posts et leurs réactions.
• PRAGMA foreign_keys = ON : doit être activé à chaque connexion SQLite, il est désactivé par défaut.


--- Requêtes complexes ---
  > Topics avec métadonnées (JOINs)
La fonction GetTopics effectue un triple JOIN (users, categories, posts, reactions) avec GROUP BY pour retourner en une seule requête : le titre, l'auteur, la catégorie, le nombre de posts et le nombre de likes de chaque topic.
  > Toggle réaction
ToggleReaction lit d'abord la réaction existante. Si elle est identique au type demandé, elle est supprimée (toggle off). Sinon, INSERT OR REPLACE remplace l'ancienne réaction ou en crée une nouvelle.
  > Filtre topics likés
Utilise une sous-requête EXISTS pour trouver les topics contenant au moins un post liké par l'utilisateur courant — évite les doublons sans DISTINCT coûteux.
  > Sujets préférés (GetFavoriteTopics)
Retourne les 10 topics où l'utilisateur a liké au moins un post OU qu'il a créé, triés par nombre de posts puis date. Utilise EXISTS et une condition OR sur t.user_id.
  > Autocomplete utilisateurs (SearchUsernames)
Requête LIKE sur le début du username (q%) avec LIMIT 8 pour les suggestions en temps réel dans le champ de réponse.
--------------------------------------------------------------------------------

================================================================================
 4. AUTHENTIFICATION & SESSIONS
================================================================================

--- Inscription ---
• Validation des champs (username, email, password non vides)
• bcrypt.GenerateFromPassword avec le coût par défaut (10 rounds)
• Insertion en BDD — les contraintes UNIQUE retournent une erreur si email ou username déjà pris
• Redirection vers /login


--- Connexion ---
• Récupération de l'utilisateur par email
• bcrypt.CompareHashAndPassword — même timing si l'utilisateur n'existe pas (pas de timing attack)
• Génération d'un UUID v4 comme identifiant de session
• Insertion en BDD avec expiration à 24h
• Cookie HttpOnly (inaccessible au JS) avec le même temps d'expiration


--- Middleware ---
Deux middlewares enveloppent les handlers :
• Auth : injecte l'utilisateur dans le contexte si un cookie valide existe, sinon injecte nil. Utilisé sur toutes les routes.
• RequireAuth : redirige vers /login si l'utilisateur est nil. Utilisé sur les routes protégées (/topic/create, /post/create, /post/react, /logout, /profile/update).


--- Nettoyage des sessions ---
Une goroutine lancée au démarrage appelle CleanExpiredSessions toutes les heures, supprimant les sessions dont expires_at est dépassé.
--------------------------------------------------------------------------------

================================================================================
 5. ROUTING
================================================================================

--- Règle critique de Go ---
http.HandleFunc matche le préfixe le plus long. Les routes spécifiques DOIVENT être déclarées avant les routes génériques, sinon elles ne sont jamais atteintes.

| Route              | Méthode  | Handler                                         |
|--------------------+----------+-------------------------------------------------|
| /                  | GET      | Home — liste des topics avec filtres            |
| /search            | GET      | Search — recherche par titre ou username        |
| /category/{slug}   | GET      | CategoryView — topics d'une catégorie           |
| /topic/create      | GET/POST | CreateTopicGET / CreateTopicPOST (protégé)      |
| /topic/{id}        | GET      | TopicView — posts d'un topic                    |
| /post/create       | POST     | CreatePost (protégé)                            |
| /post/react        | POST     | React (protégé)                                 |
| /user/{username}   | GET      | ProfileView — page profil d'un utilisateur      |
| /profile/update    | POST     | ProfileUpdate — mise à jour bio + avatar (protégé) |
| /api/users         | GET      | SearchUsers — autocomplete JSON (q=...)         |
| /login             | GET/POST | LoginGET / LoginPOST                            |
| /register          | GET/POST | RegisterGET / RegisterPOST                      |
| /logout            | GET      | Logout (protégé)                                |
| /static/           | GET      | FileServer — CSS, JS, images, avatars           |
--------------------------------------------------------------------------------

================================================================================
 6. SYSTÈME DE TEMPLATES
================================================================================

--- Problème rencontré et solution ---
Go's html/template ne supporte pas plusieurs fichiers avec le même nom de bloc ({{define "content"}}) parsés ensemble via ParseGlob. Le dernier template parsé écrase les précédents, causant l'affichage du mauvais contenu sur toutes les pages.

Solution : chaque page est compilée individuellement en combinant base.html + page.html. InitTemplates crée une map[string]*template.Template. renderTemplate cherche le template par nom dans cette map et exécute le bloc "base.html".


--- FuncMap ---
Un template.FuncMap est enregistré à la compilation avec la fonction parseMentions, qui transforme les occurrences @username en balises <span class="mention-link"> cliquables.


--- Structure des templates ---
base.html définit le layout global (header, nav, balises HTML, modal profil overlay). Chaque page redéfinit les blocs {{define "title"}} et {{define "content"}}. Le header affiche le nom de l'utilisateur connecté avec un lien vers son profil, et les boutons contextuels (connexion/déconnexion, nouveau sujet).


--- PageData ---
Une seule struct PageData est passée à tous les templates, contenant :
• User (*User) — utilisateur connecté
• ProfileUser (*User) — utilisateur dont on affiche le profil
• FavoriteTopics ([]FavoriteTopic) — sujets préférés pour la page profil
• IsOwnProfile (bool) — true si l'utilisateur consulte son propre profil
• Categories ([]Category)
• Topics ([]Topic)
• Topic (*Topic)
• Category (*Category)
• Posts ([]Post)
• Error (string)
• Filter (string)
• Search (string)
Les champs inutilisés restent à leur valeur zéro.
--------------------------------------------------------------------------------

================================================================================
 7. PROFILS UTILISATEURS
================================================================================

--- Page profil (/user/{username}) ---
Accessible à tous (visiteurs inclus). Affiche l'avatar, le nom, la date d'inscription, la bio et les sujets préférés (topics likés ou créés par l'utilisateur, limité à 10).

--- Mise à jour du profil (/profile/update) ---
Protégée par RequireAuth. Accepte un formulaire multipart avec :
• bio (champ texte libre)
• avatar (fichier image : .jpg, .jpeg, .png, .gif, .webp, max 5 Mo)

L'avatar est stocké dans static/uploads/ sous le nom avatar_{username}.{ext}. Le répertoire est créé automatiquement si absent. L'URL /static/uploads/... est enregistrée dans users.avatar_url.

--- Affichage conditionnel ---
Si l'utilisateur consulte son propre profil (IsOwnProfile = true), le formulaire d'édition est affiché. Sinon, la bio est affichée en lecture seule.
--------------------------------------------------------------------------------

================================================================================
 8. SYSTÈME DE CITATIONS & MENTIONS
================================================================================

--- Citations (posts imbriqués) ---
Chaque post peut référencer un post parent via parent_id. Dans la vue topic, un bouton "Citer" déclenche une fonction JS qui pré-remplit un aperçu de citation et renseigne le champ parent_id du formulaire de réponse. La requête GetPostsByTopic récupère le contenu du post parent et le nom de son auteur en une seule requête SQL (LEFT JOIN parent + parent_user). Si le post parent contient une image, le texte "[image]" est ajouté au contenu cité.

--- Mentions @username ---
La fonction parseMentions (template.FuncMap) détecte les occurrences @username via une regex et les transforme en <span class="mention-link" data-username="...">. Le JS associé gère l'affichage d'une modal de profil au clic sur ces spans.

--- Autocomplete ---
L'endpoint /api/users?q= retourne un tableau JSON de usernames correspondant au préfixe tapé (LIMIT 8). Le JS l'utilise pour proposer des suggestions lors de la saisie d'un @ dans le champ de réponse.
--------------------------------------------------------------------------------

================================================================================
 9. DESIGN & FRONTEND
================================================================================

--- Identité visuelle ---
| Élément          | Valeur                                          |
|------------------+-------------------------------------------------|
| Typographie      | Montserrat (Google Fonts) — 400/500/600/700/900 |
| Fond             | #0E0E0E (noir profond)                          |
| Surface          | #161616 (cartes, sidebar)                       |
| Bordures         | #2A2A2A                                         |
| Texte principal  | #F0F0F0                                         |
| Texte secondaire | #888888                                         |
| Accent principal | #7C6BFF (violet)                                |
| Like             | #4CAF50 (vert)                                  |
| Dislike          | #FF5C5C (rouge)                                 |


--- Layout ---
La page principale utilise un CSS Grid à deux colonnes : sidebar fixe de 200px + feed flexible. La sidebar devient horizontale sous 700px (media query). Le header est sticky avec backdrop-filter: blur pour un effet de profondeur.


--- Composants clés ---
• Topic card : hover avec bordure accent + translateY(-1px) pour un effet de lift subtil
• Boutons de réaction : pill avec état actif coloré (vert pour like, rouge pour dislike)
• Formulaires : fond sombre, focus avec bordure accent, labels uppercase small
• Alertes d'erreur : fond rouge semi-transparent avec bordure colorée
• Modal profil : overlay centré déclenché par le clic sur un .mention-link ou un avatar de post
• Citation preview : bloc indenté affiché au-dessus du champ de réponse avant envoi
• Avatars : image circulaire ou initiale sur fond coloré si pas d'avatar défini
--------------------------------------------------------------------------------

================================================================================
 10. FEATURES IMPLÉMENTÉES
================================================================================

--- Recherche ---
Route /search?q= avec une requête SQL LIKE sur le titre des topics ET le username de l'auteur. La barre de recherche est présente sur toutes les pages de listing (index.html). Les résultats vides affichent un état empty state.

--- Images sur les posts ---
Champ image_url (TEXT) dans la table posts. L'utilisateur colle l'URL d'une image dans le formulaire de réponse. L'image est affichée avec object-fit: cover et une hauteur max de 400px.

--- Profil utilisateur ---
Page /user/{username} avec avatar, bio, date d'inscription et sujets préférés. Upload d'avatar via formulaire multipart, stockage local dans static/uploads/.

--- Citations ---
Système de réponse imbriquée via parent_id. Bouton "Citer" sur chaque post, aperçu de la citation avant envoi, affichage du post cité dans la réponse.

--- Mentions @username ---
Détection automatique et rendu interactif des mentions. Autocomplete sur la saisie d'@ dans le champ de réponse via /api/users.
--------------------------------------------------------------------------------

================================================================================
 11. PROBLÈMES RENCONTRÉS & SOLUTIONS
================================================================================
| Problème                                      | Solution                                                                                   |
|-----------------------------------------------+--------------------------------------------------------------------------------------------|
| ParseGlob écrase les blocs define             | Compiler chaque page séparément avec ParseFiles(base.html + page.html) dans une map        |
| Nil pointer sur .Topic dans les templates     | Guard {{if .Topic}} autour de tout le contenu + redirection si ID invalide dans le handler |
| /topic/ intercepte /topic/create              | Déclarer /topic/create AVANT /topic/ dans main.go                                          |
| PRAGMA foreign_keys inactif                   | Appeler DB.Exec("PRAGMA foreign_keys = ON") après chaque sql.Open                          |
| Type mismatch Category vs Topic dans PageData | Ajouter un champ Category *Category distinct dans PageData                                 |
| Cache navigateur masquant les corrections     | Hard refresh Ctrl+Shift+R ou navigation privée                                             |
| go-sqlite3 nécessite GCC                      | Installer TDM-GCC (Windows) ou xcode-select (Mac) avant go get                             |
| parseMentions non disponible dans les templates | Enregistrer la fonction via template.FuncMap avant ParseFiles                             |
| Avatar non mis à jour si extension incorrecte | Whitelist des extensions autorisées (.jpg, .jpeg, .png, .gif, .webp) côté serveur          |
--------------------------------------------------------------------------------

================================================================================
 12. INSTALLATION & LANCEMENT
================================================================================

--- Prérequis ---
• Go 1.21+
• GCC (TDM-GCC sur Windows, xcode-select sur Mac, apt install gcc sur Linux)


--- Commandes ---
    git clone <repo> && cd forum
    go get golang.org/x/crypto/bcrypt
    go get github.com/google/uuid
    go get github.com/mattn/go-sqlite3
    go run main.go

Le serveur démarre sur http://localhost:8080. La base forum.db est créée automatiquement au premier lancement avec les 6 catégories par défaut.


--- Variables d'environnement ---
Aucune. Le port (8080) et le chemin de la BDD (./forum.db) sont codés en dur — ils peuvent être externalisés via os.Getenv si nécessaire.
--------------------------------------------------------------------------------

================================================================================
 13. AMÉLIORATIONS POSSIBLES
================================================================================
• Algorithme de score Trending : (likes - dislikes) / sqrt(age_en_heures + 2) inspiré de HackerNews
• Pagination : limiter les résultats par page avec LIMIT/OFFSET
• Mode clair/sombre : toggle CSS variables persisté en localStorage
• Modération : rôle admin pour supprimer posts et topics
• Suppression de topics/posts par leur auteur
• Validation côté client plus poussée (longueur bio, taille avatar)