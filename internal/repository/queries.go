package repository

import (
	"family-catering/internal/model"
	"fmt"
	"strings"
)

const (
	// owner's queries (owner table)
	createOwner = `
	INSERT INTO owner
		(name, email, password, phone_number)
	VALUES
		($1, $2, $3, $4)
	RETURNING id`

	updateOwner = `
	UPDATE 
		owner
	SET
		name = COALESCE(NULLIF($2, ''), name),
		phone_number = COALESCE(NULLIF($3, ''), phone_number),
		date_of_birth = COALESCE(NULLIF($4, '')::date, date_of_birth)
	WHERE id = $1`

	getOwner = `
	SELECT
		id, name, email, phone_number, date_of_birth, password 
	FROM 
		owner 
	WHERE 
		id = $1`
	getOwnerByEmail = `
	SELECT
		id, name, email, phone_number, date_of_birth, password 
		FROM 
			owner 
		WHERE 
			email = $1`
	listOwners = `SELECT
	id, name, email, phone_number, date_of_birth 
	FROM 
		owner 
	LIMIT $1 OFFSET $2`
	deleteOwner                = `DELETE FROM owner WHERE id = $1`
	updateOwnerPasswordByEmail = `UPDATE owner SET password = $2 WHERE email = $1`
	updateOwnerPasswordByID    = `UPDATE owner SET password = $2 WHERE id = $1`
	updateEmailByID            = `UPDATE owner SET email = $2 WHERE id = $1`

	// auth's queries (session table)
	insertAuthLogin = `
	INSERT INTO auth
		(sid, owner_id, email, refresh_token, jti, expired_at)
	VALUES($1, $2, $3, $4, $5, $6) RETURNING sid`

	getSessionBySessionID = `SELECT sid, owner_id, email, refresh_token, jti, expired_at FROM auth WHERE sid = $1`
	getSessionByEmail     = `SELECT sid FROM auth WHERE email = $1`
	deleteSession         = `DELETE FROM auth WHERE sid = $1`

	// menu's queries (menu table)
	getMenuByID = `
	SELECT 
		id, name, price, categories 
	FROM 
		menu 
	WHERE 
		id = $1`
	getMenuByName = `
	SELECT 
		id, name, price, categories 
	FROM 
		menu 
	WHERE 
		name = $1`
	listMenu = `
	SELECT 
		id, name, price, categories 
	FROM
		menu
	LIMIT $1 OFFSET $2`
	createMenu = `
	INSERT INTO menu
		(name, price, categories) 
	VALUES($1, $2, $3) RETURNING id`
	updateMenuByID = `
	UPDATE 
		menu 
	SET
		name = COALESCE(NULLIF($2, ''), name),
		price = COALESCE(NULLIF($3, 0), price),
		categories = COALESCE(NULLIF($4, ''), categories)
	WHERE 
		id = $1`
	deleteMenuByID = `DELETE FROM menu WHERE id = $1`

	// order's queries (order table)
	confirmPaymentViaEmail       = `UPDATE "order" SET status = 2 WHERE customer_email = $1 AND status = 1`
	updateOrderStatusToCancelled = `UPDATE "order" SET status = 3 WHERE status = 1 AND created_at > NOW() - interval '1 day' and created_at <= NOW();`
)

func menuDynamicSearchQuery(menu model.MenuQuery) (query string, args []interface{}) {
	var (
		nArgs  int
		val    string
		values []string
	)
	values = []string{}
	args = make([]interface{}, 0, 4)
	searchMenu := `SELECT id, name, price, categories FROM menu WHERE `

	if len(menu.Names) != 0 {
		var names, comparator string
		nArgs += 1
		comparator = "LIKE"
		if menu.ExactNamesMatch {
			comparator = "="
		}
		if len(menu.Names) == 1 {
			names = menu.Names[0]
			val = fmt.Sprintf(`name %s $%d`, comparator, nArgs)
		} else {
			names = strings.Join(menu.Names, ",")
			val = fmt.Sprintf(`name %s ANY(string_to_array($%d, ','))`, comparator, nArgs)
		}
		args = append(args, names)
		values = append(values, val)
	}

	if menu.Categories != "" {
		nArgs += 1
		if strings.Contains(menu.Categories, ",") {
			val = fmt.Sprintf(`string_to_array($%d, ',') && string_to_array(categories, ',')`, nArgs) // check whether lhs array is subset of rhs array
		} else {
			val = fmt.Sprintf(`$%d = ANY(string_to_array(categories, ','))`, nArgs)
		}
		values = append(values, val)
		args = append(args, menu.Categories)

	}

	if menu.MinPrice != 0 {

		nArgs += 1
		if menu.MaxPrice != 0 {
			nArgs += 1
			val = fmt.Sprintf(`price BETWEEN $%d AND $%d`, nArgs-1, nArgs)
			args = append(args, menu.MinPrice)
			args = append(args, menu.MaxPrice)
		} else {
			args = append(args, menu.MinPrice)
			val = fmt.Sprintf(`price >= $%d`, nArgs)

		}
		values = append(values, val)

	}
	if len(values) == 0 {
		return "", args
	}
	query = fmt.Sprintf("%s%s;", searchMenu, strings.Join(values, " AND "))

	return query, args
}

func dynamicSearchOrderQuery(order *model.OrderQuery) (query string, args []interface{}) {
	var (
		stmt, val string
		values    []string
		nArgs     int
	)

	stmt = `SELECT order_id, base_order_id, menu_name, customer_email,price, qty, created_at, updated_at, status FROM "order"`
	stmt = fmt.Sprintf("%s WHERE ", stmt)
	values = make([]string, 0, 8) // possible value (email, order_id, menu_names, today order, interval day, price, range price, status)
	args = make([]interface{}, 0, 8)

	if len(order.CustomerEmails) != 0 {
		var names string
		nArgs += 1
		if len(order.MenuNames) == 1 {
			names = order.MenuNames[0]
			val = fmt.Sprintf(`menu_name = $%d`, nArgs)
		} else {
			names = strings.Join(order.MenuNames, ",")
			val = fmt.Sprintf(`menu_name = ANY(string_to_array($%d, ','))`, nArgs)
		}
		args = append(args, names)
		values = append(values, val)
	}

	if len(order.MenuNames) != 0 {
		var names, comparator string
		nArgs += 1
		comparator = "LIKE"
		if order.ExactMenuNamesMatch {
			comparator = "="
		}
		if len(order.MenuNames) == 1 {
			names = order.MenuNames[0]
			val = fmt.Sprintf(`menu_name %s CONCAT('%%',$%d::TEXT,'%%')`, comparator, nArgs)
		} else {
			names = strings.Join(order.MenuNames, ",")
			val = fmt.Sprintf(`menu_name %s ANY(string_to_array($%d, ','))`, comparator, nArgs)
		}
		args = append(args, names)
		values = append(values, val)
	}

	if order.ID != 0 {
		nArgs += 1
		args = append(args, order.ID)
		val = fmt.Sprintf("order_id = $%d", nArgs)
		values = append(values, val)
	}

	if order.Status != 0 {
		nArgs += 1
		val = fmt.Sprintf("status = $%d", nArgs)
		values = append(values, val)
		args = append(args, order.Status)
	}

	if order.StartDay != "" {
		nArgs += 1
		val = fmt.Sprintf(`created_at >= $%s`, order.StartDay)
		args = append(args, order.StartDay)
		if order.EndDay != "" {
			nArgs += 1
			val = fmt.Sprintf(`%s AND created_at < $%s`, val, order.EndDay)
			args = append(args, order.EndDay)
		}

		values = append(values, val)
	}

	if order.MinPrice != 0 {
		nArgs += 1
		val = fmt.Sprintf(`price >= $%d`, nArgs)
		args = append(args, order.MinPrice)
		if order.MaxPrice != 0 {
			nArgs += 1
			val = fmt.Sprintf(`%s AND price < $%d`, val, nArgs)
			args = append(args, order.MaxPrice)
		}

		values = append(values, val)
	}

	stmt = fmt.Sprintf("%s%s", stmt, strings.Join(values, " AND "))
	return stmt, args
}
