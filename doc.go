// Package dnevnik76 doc
package dnevnik76

/* Сообщения
	URI: /messages/input/

	Страницы
	CSS Selector: #content > div.pager > span.page
	Имеет тэг a
	Если класс page_next - следующая страница
	Флаг счётчика страниц: #content > div.pager > span.page_remark Text()

	Перейти на страницу
	URI: /messages/input/?page=2

	Количество на страницу: в cookie items_perpage=[10,20,30,50]

	Список сообщений
	#content > form > table.list > tbody > tr

	Поля: Тема, От кого, Дата сообщения

	Пример сообщения
	<tr class="odd">
		<td><input type="checkbox" onclick="unselectOneCB(this, &#39;all_message_mark&#39;);" class="message_mark" name="marks" value="123456"/></td>
		<td><a href="/messages/input/123456/" class="unread">Изменение режима работы школы</a></td>
		<td>Фамилия Имя Отчество (Школа № 83, Ярославль г)</td>
           <td>17 декабря 2018 г. 18:09</td>
	</tr>

	Просмотр сообщения
	Селектор: #msgview > div.msg-text
*/

/* Marks
URI: /marks/current/

Получение для конкретной недели
	URI: /marks/current/month2/note/
	Странное поведение! Возможно получить /marks/current/month9/note/ в текущем году за сентябрь 2018
	Получается что под current имеется в виду учебный год.

	note - ученический дневник
	list - список
	date - по датам
*/
