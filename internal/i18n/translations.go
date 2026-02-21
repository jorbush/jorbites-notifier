package i18n

import "github.com/jorbush/jorbites-notifier/internal/models"

func GetUserLanguage(user *models.User) string {
	if user.Language != nil && *user.Language != "" {
		return *user.Language
	}
	return "es"
}

var emailTemplateContent = map[models.NotificationType]map[string]string{
	models.TypeNewComment: {
		"es": `
        <h2>¡Tienes un nuevo comentario!</h2>
        <p>Hola,</p>
        <p><strong>{{.Metadata.authorName}}</strong> ha dejado un comentario en tu receta.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Ver Comentario</a>
    `,
		"ca": `
        <h2>Tens un nou comentari!</h2>
        <p>Hola,</p>
        <p><strong>{{.Metadata.authorName}}</strong> ha deixat un comentari a la teva recepta.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Veure Comentari</a>
    `,
		"en": `
        <h2>You have a new comment!</h2>
        <p>Hi there,</p>
        <p><strong>{{.Metadata.authorName}}</strong> has left a comment on your recipe.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">View Comment</a>
    `,
	},
	models.TypeNewLike: {
		"es": `
        <h2>¡Alguien le ha dado like a tu receta!</h2>
        <p>Hola,</p>
        <p><strong>{{.Metadata.likedBy}}</strong> le ha dado like a tu receta.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Ver Receta</a>
    `,
		"ca": `
        <h2>Algú ha fet like a la teva recepta!</h2>
        <p>Hola,</p>
        <p><strong>{{.Metadata.likedBy}}</strong> ha fet like a la teva recepta.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Veure Recepta</a>
    `,
		"en": `
        <h2>Someone liked your recipe!</h2>
        <p>Hi there,</p>
        <p><strong>{{.Metadata.likedBy}}</strong> has liked your recipe.</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">View Recipe</a>
    `,
	},
	models.TypeNewRecipe: {
		"es": `
        <h2>¡Nueva Receta! 🍳</h2>
        <p>Hola,</p>
        <p>¡Se ha publicado una nueva receta en Jorbites!</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Ver Receta</a>
    `,
		"ca": `
        <h2>Nova Recepta! 🍳</h2>
        <p>Hola,</p>
        <p>S'ha publicat una nova recepta a Jorbites!</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Veure Recepta</a>
    `,
		"en": `
        <h2>New Recipe Alert! 🍳</h2>
        <p>Hi there,</p>
        <p>A new recipe has been posted on Jorbites!</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Check it out</a>
    `,
	},
	models.TypeNotificationsActivated: {
		"es": `
        <h2>¡Notificaciones Activadas! 🎉</h2>
        <p>Hola,</p>
        <p>Has activado con éxito las notificaciones por email de Jorbites.</p>
        <p>Ahora recibirás actualizaciones sobre:</p>
        <ul>
            <li>Nuevos comentarios en tus recetas</li>
            <li>Likes en tus recetas</li>
            <li>Nuevas recetas de tus chefs favoritos</li>
        </ul>
    `,
		"ca": `
        <h2>Notificacions Activades! 🎉</h2>
        <p>Hola,</p>
        <p>Has activat amb èxit les notificacions per email de Jorbites.</p>
        <p>Ara rebràs actualitzacions sobre:</p>
        <ul>
            <li>Nous comentaris a les teves receptes</li>
            <li>Likes a les teves receptes</li>
            <li>Noves receptes dels teus xefs favorits</li>
        </ul>
    `,
		"en": `
        <h2>Notifications Activated! 🎉</h2>
        <p>Hi there,</p>
        <p>You've successfully activated email notifications for Jorbites.</p>
        <p>You'll now receive updates about:</p>
        <ul>
            <li>New comments on your recipes</li>
            <li>Likes on your recipes</li>
            <li>New recipes from your favorite chefs</li>
        </ul>
    `,
	},
	models.TypeForgotPassword: {
		"es": `
    	<h2>Restablecer Contraseña</h2>
  		<p>Hola,</p>
  		<p>Has solicitado restablecer tu contraseña. Haz clic en el siguiente enlace para crear una nueva contraseña:</p>
  		<a href="{{.Metadata.resetUrl}}" class="button">Restablecer Contraseña</a>
    	<p>Este enlace expirará en 1 hora.</p>
        <p>Si no solicitaste este cambio, puedes ignorar este correo.</p>
    `,
		"ca": `
    	<h2>Restablir Contrasenya</h2>
  		<p>Hola,</p>
  		<p>Has sol·licitat restablir la teva contrasenya. Fes clic a l'enllaç següent per crear una nova contrasenya:</p>
  		<a href="{{.Metadata.resetUrl}}" class="button">Restablir Contrasenya</a>
    	<p>Aquest enllaç expirarà en 1 hora.</p>
        <p>Si no has sol·licitat aquest canvi, pots ignorar aquest correu.</p>
    `,
		"en": `
    	<h2>Password Reset</h2>
  		<p>Hi there,</p>
  		<p>You have requested to reset your password. Click on the following link to create a new password:</p>
  		<a href="{{.Metadata.resetUrl}}" class="button">Reset Password</a>
    	<p>This link will expire in 1 hour.</p>
        <p>If you did not request this change, you can ignore this email.</p>
    `,
	},
	models.TypeMentionInComment: {
		"es": `
    	<h2>¡Te mencionaron en un comentario!</h2>
  		<p>Hola,</p>
    	<p><strong>{{.Metadata.authorName}}</strong> te mencionó en un comentario de una receta.</p>
    	<p>Haz clic en el botón de abajo para ver la receta:</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Ver Receta</a>
    `,
		"ca": `
    	<h2>T'han mencionat en un comentari!</h2>
  		<p>Hola,</p>
    	<p><strong>{{.Metadata.authorName}}</strong> t'ha mencionat en un comentari d'una recepta.</p>
    	<p>Fes clic al botó de sota per veure la recepta:</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">Veure Recepta</a>
    `,
		"en": `
    	<h2>You were mentioned in a comment!</h2>
  		<p>Hi there,</p>
    	<p><strong>{{.Metadata.authorName}}</strong> mentioned you in a comment on a recipe.</p>
    	<p>Click the button below to view the recipe:</p>
        <a href="{{.SiteURL}}/recipes/{{.Metadata.recipeId}}" class="button">View Recipe</a>
    `,
	},
	models.TypeNewBlog: {
		"es": `
		<h2>¡Nueva Entrada de Blog! 📝</h2>
		<p>Hola,</p>
		<p>¡Se ha publicado una nueva entrada de blog en Jorbites!</p>
		<a href="{{.SiteURL}}/blog/{{.Metadata.blog_id}}" class="button">Leer Ahora</a>
	`,
		"ca": `
		<h2>Nova Entrada de Blog! 📝</h2>
		<p>Hola,</p>
		<p>S'ha publicat una nova entrada de blog a Jorbites!</p>
		<a href="{{.SiteURL}}/blog/{{.Metadata.blog_id}}" class="button">Llegir Ara</a>
	`,
		"en": `
		<h2>New Blog Post! 📝</h2>
		<p>Hi there,</p>
		<p>A new blog post has been published on Jorbites!</p>
		<a href="{{.SiteURL}}/blog/{{.Metadata.blog_id}}" class="button">Read it now</a>
	`,
	},
	models.TypeNewEvent: {
		"es": `
		<h2>¡Nuevo Evento! 🎉</h2>
		<p>Hola,</p>
		<p>¡Se ha publicado un nuevo evento en Jorbites!</p>
		<a href="{{.SiteURL}}/events/{{.Metadata.eventId}}" class="button">Ver Evento</a>
	`,
		"ca": `
		<h2>Nou Esdeveniment! 🎉</h2>
		<p>Hola,</p>
		<p>S'ha publicat un nou esdeveniment a Jorbites!</p>
		<a href="{{.SiteURL}}/events/{{.Metadata.eventId}}" class="button">Veure Esdeveniment</a>
	`,
		"en": `
		<h2>New Event! 🎉</h2>
		<p>Hi there,</p>
		<p>A new event has been published on Jorbites!</p>
		<a href="{{.SiteURL}}/events/{{.Metadata.eventId}}" class="button">View Event</a>
	`,
	},
	models.TypeEventEndingSoon: {
		"es": `
		<h2>¡El Evento Está por Terminar! ⏰</h2>
		<p>Hola,</p>
		<p>El evento <strong>{{.Metadata.title}}</strong> terminará en 3 días.</p>
		<p>¡No te lo pierdas! Todavía tienes tiempo para participar.</p>
		<a href="{{.SiteURL}}/events/{{.Metadata.eventId}}" class="button">Ver Evento</a>
	`,
		"ca": `
		<h2>L'Esdeveniment Està a Punt d'Acabar! ⏰</h2>
		<p>Hola,</p>
		<p>L'esdeveniment <strong>{{.Metadata.title}}</strong> acabarà en 3 dies.</p>
		<p>No te'l perdis! Encara tens temps per participar.</p>
		<a href="{{.SiteURL}}/events/{{.Metadata.eventId}}" class="button">Veure Esdeveniment</a>
	`,
		"en": `
		<h2>Event Ending Soon! ⏰</h2>
		<p>Hi there,</p>
		<p>The event <strong>{{.Metadata.title}}</strong> will end in 3 days.</p>
		<p>Don't miss out! You still have time to participate.</p>
		<a href="{{.SiteURL}}/events/{{.Metadata.eventId}}" class="button">View Event</a>
	`,
	},
	models.TypeNewQuest: {
		"es": `
		<h2>¡Nueva Misión! 🎯</h2>
		<p>Hola,</p>
		<p>¡Alguien ha solicitado una nueva receta en Jorbites!</p>
		<a href="{{.SiteURL}}/quests/{{.Metadata.questId}}" class="button">Ver Misión</a>
	`,
		"ca": `
		<h2>Nova Missió! 🎯</h2>
		<p>Hola,</p>
		<p>Algú ha sol·licitat una nova recepta a Jorbites!</p>
		<a href="{{.SiteURL}}/quests/{{.Metadata.questId}}" class="button">Veure Missió</a>
	`,
		"en": `
		<h2>New Quest! 🎯</h2>
		<p>Hi there,</p>
		<p>Someone has requested a new recipe on Jorbites!</p>
		<a href="{{.SiteURL}}/quests/{{.Metadata.questId}}" class="button">View Quest</a>
	`,
	},
	models.TypeQuestFulfilled: {
		"es": `
		<h2>¡Tu Misión ha sido completada! 🏆</h2>
		<p>Hola,</p>
		<p><strong>{{.Metadata.fulfilledByName}}</strong> ha completado tu misión con una receta.</p>
		<a href="{{.SiteURL}}/quests/{{.Metadata.questId}}" class="button">Ver Envío</a>
	`,
		"ca": `
		<h2>La teva Missió ha estat completada! 🏆</h2>
		<p>Hola,</p>
		<p><strong>{{.Metadata.fulfilledByName}}</strong> ha completat la teva missió amb una recepta.</p>
		<a href="{{.SiteURL}}/quests/{{.Metadata.questId}}" class="button">Veure Enviament</a>
	`,
		"en": `
		<h2>Your Quest has been fulfilled! 🏆</h2>
		<p>Hi there,</p>
		<p><strong>{{.Metadata.fulfilledByName}}</strong> has fulfilled your quest with a recipe.</p>
		<a href="{{.SiteURL}}/quests/{{.Metadata.questId}}" class="button">View Submission</a>
	`,
	},
}

var emailSubjects = map[models.NotificationType]map[string]string{
	models.TypeNewComment: {
		"es": "Nuevo Comentario en tu Receta - Jorbites",
		"ca": "Nou Comentari a la teva Recepta - Jorbites",
		"en": "New Comment on Your Recipe - Jorbites",
	},
	models.TypeNewLike: {
		"es": "Nuevo Like en tu Receta - Jorbites",
		"ca": "Nou Like a la teva Recepta - Jorbites",
		"en": "New Like on Your Recipe - Jorbites",
	},
	models.TypeNewRecipe: {
		"es": "Nueva Receta Disponible - Jorbites",
		"ca": "Nova Recepta Disponible - Jorbites",
		"en": "New Recipe Available - Jorbites",
	},
	models.TypeNotificationsActivated: {
		"es": "Bienvenido a las Notificaciones de Jorbites",
		"ca": "Benvingut a les Notificacions de Jorbites",
		"en": "Welcome to Jorbites Notifications",
	},
	models.TypeForgotPassword: {
		"es": "Solicitud de Restablecimiento de Contraseña - Jorbites",
		"ca": "Sol·licitud de Restabliment de Contrasenya - Jorbites",
		"en": "Password Reset Request - Jorbites",
	},
	models.TypeMentionInComment: {
		"es": "Te Mencionaron en un Comentario - Jorbites",
		"ca": "T'han Mencionat en un Comentari - Jorbites",
		"en": "You Were Mentioned in a Comment - Jorbites",
	},
	models.TypeNewBlog: {
		"es": "Nueva Entrada de Blog Disponible - Jorbites",
		"ca": "Nova Entrada de Blog Disponible - Jorbites",
		"en": "New Blog Post Available - Jorbites",
	},
	models.TypeNewEvent: {
		"es": "Nuevo Evento Disponible - Jorbites",
		"ca": "Nou Esdeveniment Disponible - Jorbites",
		"en": "New Event Available - Jorbites",
	},
	models.TypeEventEndingSoon: {
		"es": "¡El Evento Está por Terminar! - Jorbites",
		"ca": "L'Esdeveniment Està a Punt d'Acabar! - Jorbites",
		"en": "Event Ending Soon! - Jorbites",
	},
	models.TypeNewQuest: {
		"es": "Nueva Misión Disponible - Jorbites",
		"ca": "Nova Missió Disponible - Jorbites",
		"en": "New Quest Available - Jorbites",
	},
	models.TypeQuestFulfilled: {
		"es": "¡Tu Misión ha sido completada! - Jorbites",
		"ca": "La teva Missió ha estat completada! - Jorbites",
		"en": "Your Quest has been fulfilled! - Jorbites",
	},
}

var baseTemplateFooter = map[string]string{
	"es": `
            <p>Estás recibiendo este correo porque tienes las notificaciones activadas en Jorbites.</p>
            <p>Para gestionar tus preferencias de correo, ve a <a href="{{.SiteURL}}">Ajustes → Notificaciones por Email</a></p>
            <p>© {{.CurrentYear}} Jorbites. Todos los derechos reservados.</p>
    `,
	"ca": `
            <p>Estàs rebent aquest correu perquè tens les notificacions activades a Jorbites.</p>
            <p>Per gestionar les teves preferències de correu, vés a <a href="{{.SiteURL}}">Configuració → Notificacions per Email</a></p>
            <p>© {{.CurrentYear}} Jorbites. Tots els drets reservats.</p>
    `,
	"en": `
            <p>You're receiving this email because you have notifications enabled on Jorbites.</p>
            <p>To manage your email preferences, go to <a href="{{.SiteURL}}">Settings → Email Notifications</a></p>
            <p>© {{.CurrentYear}} Jorbites. All rights reserved.</p>
    `,
}

// GetEmailTemplateContent returns the email template content for a given notification type and language
func GetEmailTemplateContent(notificationType models.NotificationType, language string) string {
	templates, exists := emailTemplateContent[notificationType]
	if !exists {
		return ""
	}

	content, exists := templates[language]
	if !exists {
		// Fallback to Spanish
		return templates["es"]
	}

	return content
}

func GetEmailSubject(notificationType models.NotificationType, language string) string {
	subjects, exists := emailSubjects[notificationType]
	if !exists {
		if language == "ca" {
			return "Notificació de Jorbites"
		} else if language == "en" {
			return "Notification from Jorbites"
		}
		return "Notificación de Jorbites"
	}

	subject, exists := subjects[language]
	if !exists {
		return subjects["es"]
	}

	return subject
}

func GetBaseTemplateFooter(language string) string {
	footer, exists := baseTemplateFooter[language]
	if !exists {
		return baseTemplateFooter["es"]
	}
	return footer
}

type PushNotificationTexts struct {
	Title   string
	Message string
}

func GetPushNotificationText(notificationType models.NotificationType, language string, metadata map[string]string) PushNotificationTexts {
	switch notificationType {
	case models.TypeNewLike:
		likedBy := metadata["likedBy"]
		switch language {
		case "ca":
			if likedBy != "" {
				return PushNotificationTexts{Title: "Nou Like", Message: likedBy + " ha fet like a la teva recepta"}
			}
			return PushNotificationTexts{Title: "Nou Like", Message: "Algú ha fet like a la teva recepta"}
		case "en":
			if likedBy != "" {
				return PushNotificationTexts{Title: "New Like", Message: likedBy + " liked your recipe"}
			}
			return PushNotificationTexts{Title: "New Like", Message: "Someone liked your recipe"}
		default: // es
			if likedBy != "" {
				return PushNotificationTexts{Title: "Nuevo Like", Message: likedBy + " le ha dado like a tu receta"}
			}
			return PushNotificationTexts{Title: "Nuevo Like", Message: "Alguien le ha dado like a tu receta"}
		}

	case models.TypeNewComment:
		authorName := metadata["authorName"]
		switch language {
		case "ca":
			if authorName != "" {
				return PushNotificationTexts{Title: "Nou Comentari", Message: authorName + " ha comentat a la teva recepta"}
			}
			return PushNotificationTexts{Title: "Nou Comentari", Message: "Nou comentari a la teva recepta"}
		case "en":
			if authorName != "" {
				return PushNotificationTexts{Title: "New Comment", Message: authorName + " commented on your recipe"}
			}
			return PushNotificationTexts{Title: "New Comment", Message: "New comment on your recipe"}
		default: // es
			if authorName != "" {
				return PushNotificationTexts{Title: "Nuevo Comentario", Message: authorName + " ha comentado en tu receta"}
			}
			return PushNotificationTexts{Title: "Nuevo Comentario", Message: "Nuevo comentario en tu receta"}
		}

	case models.TypeNotificationsActivated:
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "Notificacions Activades", Message: "Has activat les notificacions amb èxit"}
		case "en":
			return PushNotificationTexts{Title: "Notifications Activated", Message: "You have successfully activated notifications"}
		default: // es
			return PushNotificationTexts{Title: "Notificaciones Activadas", Message: "Has activado las notificaciones con éxito"}
		}

	case models.TypeMentionInComment:
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "T'han mencionat!", Message: "T'han mencionat en un comentari"}
		case "en":
			return PushNotificationTexts{Title: "You were mentioned!", Message: "You were mentioned in a comment"}
		default: // es
			return PushNotificationTexts{Title: "¡Te mencionaron!", Message: "Te mencionaron en un comentario"}
		}

	case models.TypeNewRecipe:
		recipeName := metadata["recipeName"]
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "Nova Recepta!", Message: "Nova recepta disponible: " + recipeName}
		case "en":
			return PushNotificationTexts{Title: "New Recipe!", Message: "New recipe available: " + recipeName}
		default: // es
			return PushNotificationTexts{Title: "¡Nueva Receta!", Message: "Nueva receta disponible: " + recipeName}
		}

	case models.TypeNewBlog:
		postTitle := metadata["title"]
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "Nou Post de Blog!", Message: postTitle}
		case "en":
			return PushNotificationTexts{Title: "New Blog Post!", Message: postTitle}
		default: // es
			return PushNotificationTexts{Title: "¡Nuevo Post de Blog!", Message: postTitle}
		}

	case models.TypeNewEvent:
		eventTitle := metadata["title"]
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "Nou Esdeveniment!", Message: eventTitle}
		case "en":
			return PushNotificationTexts{Title: "New Event!", Message: eventTitle}
		default: // es
			return PushNotificationTexts{Title: "¡Nuevo Evento!", Message: eventTitle}
		}

	case models.TypeEventEndingSoon:
		eventTitle := metadata["title"]
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "L'Esdeveniment Acaba Aviat!", Message: eventTitle + " acaba en 3 dies"}
		case "en":
			return PushNotificationTexts{Title: "Event Ending Soon!", Message: eventTitle + " ends in 3 days"}
		default: // es
			return PushNotificationTexts{Title: "¡Evento Terminando!", Message: eventTitle + " termina en 3 días"}
		}

	case models.TypeNewQuest:
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "Nova Quest!", Message: "S'ha publicat una nova quest a Jorbites"}
		case "en":
			return PushNotificationTexts{Title: "New Quest!", Message: "A new quest has been posted on Jorbites"}
		default: // es
			return PushNotificationTexts{Title: "¡Nueva Quest!", Message: "Se ha publicado una nueva quest en Jorbites"}
		}

	case models.TypeQuestFulfilled:
		fulfilledByName := metadata["fulfilledByName"]
		switch language {
		case "ca":
			if fulfilledByName != "" {
				return PushNotificationTexts{Title: "Missió Completada!", Message: fulfilledByName + " ha completat la teva missió"}
			}
			return PushNotificationTexts{Title: "Missió Completada!", Message: "La teva missió ha estat completada"}
		case "en":
			if fulfilledByName != "" {
				return PushNotificationTexts{Title: "Quest Fulfilled!", Message: fulfilledByName + " fulfilled your quest"}
			}
			return PushNotificationTexts{Title: "Quest Fulfilled!", Message: "Your quest has been fulfilled"}
		default: // es
			if fulfilledByName != "" {
				return PushNotificationTexts{Title: "¡Misión Completada!", Message: fulfilledByName + " ha completado tu misión"}
			}
			return PushNotificationTexts{Title: "¡Misión Completada!", Message: "Tu misión ha sido completada"}
		}

	default:
		switch language {
		case "ca":
			return PushNotificationTexts{Title: "Notificació", Message: "Tens una nova notificació"}
		case "en":
			return PushNotificationTexts{Title: "Notification", Message: "You have a new notification"}
		default: // es
			return PushNotificationTexts{Title: "Notificación", Message: "Tienes una nueva notificación"}
		}
	}
}
