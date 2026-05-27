import os

# Liste des fichiers essentiels pour l'audit de sécurité
ESSENTIAL_FILES = [
    "main.go",
    "database/db.go",
    "database/schema.sql",
    "handlers/auth.go",
    "handlers/helpers.go",
    "handlers/posts.go",
    "handlers/profile.go",
    "handlers/topics.go",
    "middleware/auth.go",
    "models/models.go",
    "templates/base.html",
    "templates/create_topic.html",
    "templates/index.html",
    "templates/login.html",
    "templates/profile.html",
    "templates/register.html",
    "templates/topic.html"
]

OUTPUT_FILENAME = "code_pour_audit.txt"

def export_essential_code():
    count = 0
    with open(OUTPUT_FILENAME, "w", encoding="utf-8") as outfile:
        for filepath in ESSENTIAL_FILES:
            if os.path.exists(filepath):
                outfile.write(f"// ==========================================\n")
                outfile.write(f"// FILE: {filepath}\n")
                outfile.write(f"// ==========================================\n\n")
                try:
                    with open(filepath, "r", encoding="utf-8") as infile:
                        outfile.write(infile.read())
                    count += 1
                except Exception as e:
                    outfile.write(f"// Erreur lors de la lecture du fichier : {e}\n")
                outfile.write("\n\n\n")
            else:
                print(f"[-] Fichier introuvable : {filepath}")
                
    print(f"[+] Terminé ! {count} fichiers ont été regroupés dans '{OUTPUT_FILENAME}'.")

if __name__ == "__main__":
    export_essential_code()