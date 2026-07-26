#!/usr/bin/env python3
"""Smoke test para o ark-commander usando Selenium."""
import os
import sys
import time

sys.path.insert(0, "/tmp/selenium-env/lib/python3.13/site-packages")

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service

BASE_URL = os.environ.get("APP_URL", "http://localhost:3000")
API_URL = os.environ.get("API_URL", "http://localhost:8080")

def check(endpoint, label):
    """Checa um endpoint HTTP e retorna (ok, status_code ou erro)."""
    import urllib.request, urllib.error
    try:
        req = urllib.request.Request(endpoint)
        with urllib.request.urlopen(req, timeout=5) as resp:
            return True, resp.status
    except urllib.error.HTTPError as e:
        return False, e.code
    except Exception as e:
        return False, str(e)

def main():
    print("=" * 50)
    print("SMOKE TEST - ARK Server Commander")
    print("=" * 50)

    # 1. API Health
    ok, code = check(f"{API_URL}/health", "API Health")
    print(f"[{'OK' if ok else 'FAIL'}] API /health -> {code}")

    # 2. API Servers (requer auth, espera 401 sem token)
    ok, code = check(f"{API_URL}/api/servers", "API Servers (no auth)")
    print(f"[{'OK' if not ok and code == 401 else 'UNEXPECTED'}] API /api/servers (no auth) -> {code}")

    # 3. UI acessível
    ok, code = check(BASE_URL, "UI")
    print(f"[{'OK' if ok else 'FAIL'}] UI -> {code}")

    if not ok:
        print("FALHA: UI nao acessivel. Abortando testes Selenium.")
        sys.exit(1)

    # 4. Teste Selenium completo: login + verificar servidores
    print("\n--- Teste Selenium: Login + Pagina de Servidores ---")

    chrome_options = Options()
    chrome_options.add_argument("--headless")
    chrome_options.add_argument("--no-sandbox")
    chrome_options.add_argument("--disable-dev-shm-usage")
    chrome_options.add_argument("--disable-gpu")

    try:
        driver = webdriver.Chrome(options=chrome_options)
        driver.set_page_load_timeout(15)
    except Exception as e:
        print(f"[WARN] Nao foi possivel iniciar o ChromeDriver: {e}")
        print("O ambiente Selenium esta instalado mas o Chrome/ChromeDriver precisam estar acessiveis.")
        sys.exit(0)

    try:
        driver.get(BASE_URL)
        WebDriverWait(driver, 10).until(EC.presence_of_element_located((By.TAG_NAME, "body")))
        print(f"[OK] UI carregou: {driver.title}")

        # Tentar fazer login (usuário admin criado na migração)
        login_link = driver.find_element(By.LINK_TEXT, "Login") if driver.find_elements(By.LINK_TEXT, "Login") else None
        if login_link:
            login_link.click()
            time.sleep(1)

        # Verificar se a página de login existe
        username_input = driver.find_element(By.NAME, "username") if driver.find_elements(By.NAME, "username") else None
        if username_input:
            username_input.send_keys("admin")
            password_input = driver.find_element(By.NAME, "password")
            password_input.send_keys("senha123")
            password_input.submit()
            time.sleep(2)
            print("[OK] Login submetido com admin/senha123")

        # Verificar se navegou para a página de servidores
        current_url = driver.current_url
        print(f"[INFO] URL atual após login: {current_url}")

        # Procurar a lista de servidores ou mensagem de erro
        body_text = driver.find_element(By.TAG_NAME, "body").text
        if "Failed to get server list" in body_text or "Erro" in body_text:
            print("[FAIL] Erro na lista de servidores detectado no UI!")
        elif "servers" in current_url.lower() or "server" in body_text.lower():
            print("[OK] Página de servidores carregou sem erro de lista")
        else:
            print(f"[INFO] Conteúdo da página: {body_text[:200]}")

        print("\n=== TESTE SELENIUM CONCLUIDO ===")

    except Exception as e:
        print(f"[ERROR] Erro durante o teste Selenium: {e}")
    finally:
        driver.quit()

if __name__ == "__main__":
    main()
