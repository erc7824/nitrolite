import { Users } from "lucide-react";
import { Badge } from "./ui/badge";
import { cn } from "../lib/utils";
import { useEffect } from "react";

interface OnlinePlayersCounterProps {
    className?: string;
    count?: number;
}

export function OnlinePlayersCounter({ className, count = 1 }: OnlinePlayersCounterProps) {
    // For debugging - log count when it changes
    useEffect(() => {
        console.log("OnlinePlayersCounter rendering with count:", count);
    }, [count]);

    return (
        <div className={cn("flex items-center", className)}>
            <Badge variant="secondary" className="bg-viper-purple/30 hover:bg-viper-purple/40 text-viper-purple-light border-viper-purple/30">
                <Users className="h-3 w-3 mr-1.5" />
                <span>{count} online</span>
            </Badge>
        </div>
    );
}
