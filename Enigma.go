package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

// Rotor структура для роторов
type Rotor struct {
	mapping  string // Отображение ротора (последовательность букв)
	notch    byte   // Символ на которых ротор поварачивается
	position int    // Текущая позиция ротора (от 0 до 25)
}

// Plugboard структура для патч-панели
type Plugboard map[byte]byte

// Reflector строка для отражателя
type Reflector string

// Преобразование буквы через патч-панель
func (p Plugboard) Transform(letter byte) byte {
	if transformed, ok := p[letter]; ok { // Проверяем, есть ли замена для буквы в патч-панели
		return transformed
	}
	return letter
}

// Преобразование буквы через ротор в прямом направлении (при шифровании)
func (r *Rotor) TransformForward(letter byte) byte {
	// Сдвиг позиции буквы относительно ротора
	offset := (int(letter-'A') + r.position) % 26 // Вычисляем смещение буквы с учетом позиции ротора
	transformed := r.mapping[offset]              // Получаем замененную букву из отображения ротора
	// Возврат с учётом обратного сдвига
	return byte((int(transformed-'A')-r.position+26)%26 + 'A') // Возвращаем букву, смещенную обратно к исходной позиции
}

// Преобразование буквы через ротор в обратном направлении (при расшифровке)
func (r *Rotor) TransformBackward(letter byte) byte {
	// Сдвиг позиции буквы относительно ротора
	shifted := (int(letter-'A') + r.position) % 26 // Вычисляем смещение буквы с учетом позиции ротора
	// Находим индекс в роторе
	index := strings.IndexByte(r.mapping, byte(shifted+'A')) // Ищем позицию буквы в отображении ротора
	// Возврат с учётом обратного сдвига
	return byte((index-r.position+26)%26 + 'A') // Возвращаем букву, смещенную обратно к исходной позиции
}

// Сдвиг ротора на одну позицию
func (r *Rotor) Rotate() bool {
	r.position = (r.position + 1) % 26      // Увеличиваем позицию ротора, обнуляя при достижении 26
	return r.mapping[r.position] == r.notch // Возвращаем true, если текущая позиция ротора соответствует выемке
}

// Основная функция шифрования одной буквы
func EncryptLetter(letter byte, rotors []*Rotor, reflector Reflector, plugboard Plugboard) byte {
	// Преобразование через патч-панель
	letter = plugboard.Transform(letter)

	// Проход через роторы (вперёд)
	for _, rotor := range rotors {
		letter = rotor.TransformForward(letter) // Применяем прямое преобразование ротора
	}

	// Преобразование через отражатель
	letter = reflector[letter-'A']

	// Проход через роторы (обратно)
	for i := len(rotors) - 1; i >= 0; i-- { // Проходим через каждый ротор в обратном направлении
		letter = rotors[i].TransformBackward(letter)
	}

	// Преобразование через патч-панель (обратно)
	return plugboard.Transform(letter)
}

// Функция сдвиг всех роторов, начиная с правого (быстрого)
func RotateRotors(rotors []*Rotor) {
	// Начинаем с быстродействующего ротора
	rotateNext := rotors[0].Rotate()                 // Сдвигаем первый ротор, получаем флаг, нужно ли сдвигать следующий
	for i := 1; i < len(rotors) && rotateNext; i++ { // Проходим по остальным роторам, пока есть необходимость сдвигать
		rotateNext = rotors[i].Rotate() // Сдвигаем ротор, флаг определяет, нужно ли сдвигать следующий
	}
}

// Функция шифрование всего сообщения
func EncryptMessage(message string, rotors []*Rotor, reflector Reflector, plugboard Plugboard) string {
	encrypted := ""                  // Строка для накопления зашифрованного сообщения
	for _, letter := range message { // Итерируем по каждой букве (руне) в сообщении
		if letter >= 'A' && letter <= 'Z' { // Проверяем, является ли буква английской заглавной
			RotateRotors(rotors)                                                           // Сдвигаем роторы перед шифрованием буквы
			encrypted += string(EncryptLetter(byte(letter), rotors, reflector, plugboard)) // Шифруем букву и добавляем к результату
		} else {
			encrypted += string(letter) // Если не английская заглавная, добавляем символ без изменений
		}
	}
	return encrypted
}

func main() {
	// Настройка роторов
	rotors := []*Rotor{
		{mapping: "EKMFLGDQVZNTOWYHXUSPAIBRCJ", notch: 'Q', position: 0},
		{mapping: "AJDKSIRUXBLHWTMCQGZNPYFVOE", notch: 'E', position: 0},
		{mapping: "BDFHJLCPRTXVZNYEIWGAKMUSQO", notch: 'V', position: 0},
	}

	// Настройка отражателя
	reflector := Reflector("YRUHQSLDPXNGOKMIEBFZCWVJAT")

	// Настройка патч-панели
	plugboard := Plugboard{
		'A': 'Z', 'Z': 'A', //Замена A <-> Z
		'B': 'Y', 'Y': 'B', //Замена B <-> Y
	}

	// Чтение сообщения из файла text.txt
	content, err := ioutil.ReadFile("text.txt")
	if err != nil {
		log.Fatalf("Не удалось прочитать файл text.txt: %v", err)
		return
	}
	message := string(content)         // Преобразуем содержимое файла в строку
	message = strings.ToUpper(message) // Приводим сообщение к верхнему регистру для шифрования только английских букв

	// Шифрование сообщения с таймером
	startTime := time.Now() // Засекаем время начала шифрования
	encrypted := EncryptMessage(message, rotors, reflector, plugboard)
	encryptionTime := time.Since(startTime) // Вычисляем время шифрования

	// Запись зашифрованного сообщения в файл encrypted.txt
	err = ioutil.WriteFile("encrypted.txt", []byte(encrypted), 0644)
	if err != nil {
		log.Fatalf("Не удалось записать зашифрованное сообщение в файл encrypted.txt: %v", err)
		return
	}
	fmt.Println("Зашифрованное сообщение записано в encrypted.txt")

	// Сброс позиций роторов для расшифровки
	for _, rotor := range rotors {
		rotor.position = 0
	}

	// Расшифровка сообщения с таймером
	startTime = time.Now() // Засекаем время начала расшифровки
	decrypted := EncryptMessage(encrypted, rotors, reflector, plugboard)
	decryptionTime := time.Since(startTime) // Вычисляем время расшифровки

	// Запись расшифрованного сообщения в файл decrypted.txt
	err = ioutil.WriteFile("decrypted.txt", []byte(decrypted), 0644)
	if err != nil {
		log.Fatalf("Не удалось записать расшифрованное сообщение в файл decrypted.txt: %v", err)
		return
	}
	fmt.Println("Расшифрованное сообщение записано в decrypted.txt")

	// Вывод времени выполнения
	fmt.Printf("Время шифрования: %v\n", encryptionTime)
	fmt.Printf("Время расшифровки: %v\n", decryptionTime)
}
