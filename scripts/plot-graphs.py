import json
import matplotlib.pyplot as plt
from datetime import datetime
import os

# Создаём папку для графиков, если её нет
os.makedirs('results/graphs', exist_ok=True)

def load_metrics(filename):
    """Загружает данные из JSON-файла k6"""
    times = []
    values = []
    
    if not os.path.exists(filename):
        print(f"Файл {filename} не найден")
        return times, values
    
    with open(filename, 'r', encoding='utf-8') as f:
        for line in f:
            try:
                data = json.loads(line.strip())
                if data.get('metric') == 'http_req_duration':
                    point = data.get('data', {})
                    time_str = point.get('time', '')
                    if time_str:
                        try:
                            t = datetime.fromisoformat(time_str.replace('Z', '+00:00'))
                            times.append(t)
                            values.append(point.get('value', 0))
                        except:
                            pass
            except json.JSONDecodeError:
                continue
    
    return times, values

def plot_smoke_test(times, values):
    """График для smoke-теста"""
    if not times:
        return
    
    fig, ax = plt.subplots(figsize=(10, 6))
    
    ax.plot(range(1, len(values)+1), values, 'o-', color='blue', linewidth=2, markersize=8)
    ax.axhline(y=sum(values)/len(values), color='red', linestyle='--', label=f'Среднее: {sum(values)/len(values):.2f} мс')
    
    ax.set_title('Smoke Test - Время ответа по запросам', fontsize=14)
    ax.set_xlabel('Номер запроса', fontsize=12)
    ax.set_ylabel('Время ответа (мс)', fontsize=12)
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    plt.tight_layout()
    plt.savefig('results/graphs/smoke-test.png', dpi=150)
    plt.show()
    print("График smoke-теста сохранён: results/graphs/smoke-test.png")

def plot_baseline(times, values, title, filename):
    """График для baseline-теста"""
    if not times:
        print(f"Нет данных для {title}")
        return
    
    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))
    
    # График 1: Время ответа по времени
    ax1.plot(times, values, 'o-', color='green', markersize=1, linewidth=0.5)
    ax1.axhline(y=sum(values)/len(values), color='red', linestyle='--', label=f'Среднее: {sum(values)/len(values):.2f} мс')
    ax1.set_title(f'{title} - Время ответа по времени', fontsize=14)
    ax1.set_xlabel('Время', fontsize=12)
    ax1.set_ylabel('Время ответа (мс)', fontsize=12)
    ax1.grid(True, alpha=0.3)
    ax1.legend()
    
    # График 2: Гистограмма распределения
    ax2.hist(values, bins=50, color='green', alpha=0.7, edgecolor='black')
    ax2.set_title(f'{title} - Распределение времени ответа', fontsize=14)
    ax2.set_xlabel('Время ответа (мс)', fontsize=12)
    ax2.set_ylabel('Количество запросов', fontsize=12)
    ax2.grid(True, alpha=0.3)
    
    plt.tight_layout()
    plt.savefig(f'results/graphs/{filename}.png', dpi=150)
    plt.show()
    print(f"График {title} сохранён: results/graphs/{filename}.png")

def compare_baseline(data1, data2, label1, label2):
    """Сравнительный график двух тестов"""
    if not data1 or not data2:
        print("Недостаточно данных для сравнения")
        return
    
    fig, ax = plt.subplots(figsize=(12, 6))
    
    # Строим гистограммы
    ax.hist(data1, bins=50, alpha=0.5, label=label1, color='blue')
    ax.hist(data2, bins=50, alpha=0.5, label=label2, color='orange')
    
    ax.set_title('Сравнение времени ответа: без прокси vs с прокси', fontsize=14)
    ax.set_xlabel('Время ответа (мс)', fontsize=12)
    ax.set_ylabel('Количество запросов', fontsize=12)
    ax.legend()
    ax.grid(True, alpha=0.3)
    
    plt.tight_layout()
    plt.savefig('results/graphs/comparison.png', dpi=150)
    plt.show()
    print("Сравнительный график сохранён: results/graphs/comparison.png")

def main():
    print("Загружаем данные...")
    
    # Загружаем данные из всех трёх тестов
    smoke_times, smoke_values = load_metrics('results/smoke-test.json')
    baseline_times, baseline_values = load_metrics('results/baseline-without-proxy.json')
    proxy_times, proxy_values = load_metrics('results/baseline-with-proxy.json')
    
    print(f"Smoke-тест: {len(smoke_values)} записей")
    print(f"Baseline без прокси: {len(baseline_values)} записей")
    print(f"Baseline с прокси: {len(proxy_values)} записей")
    
    # Строим графики
    if smoke_values:
        plot_smoke_test(smoke_times, smoke_values)
    else:
        print("Нет данных для smoke-теста")
    
    if baseline_values:
        plot_baseline(baseline_times, baseline_values, "Baseline без прокси (500 RPS)", "baseline-без-прокси")
    else:
        print("Нет данных для baseline-теста без прокси")
    
    if proxy_values:
        plot_baseline(proxy_times, proxy_values, "Baseline с прокси (500 RPS)", "baseline-с-прокси")
    else:
        print("Нет данных для baseline-теста с прокси")
    
    # Сравнительный график
    if baseline_values and proxy_values:
        compare_baseline(baseline_values, proxy_values, "Без прокси", "С прокси")
    
    print("\nВсе графики сохранены в папке results/graphs/")

if __name__ == "__main__":
    main()